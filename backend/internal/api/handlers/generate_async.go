package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/omnigen/backend/internal/adapters"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/service"
	"go.uber.org/zap"
)

// ClipVideo represents a generated video clip
type ClipVideo struct {
	VideoURL     string
	LastFrameURL string
	Duration     float64
}

// S3 key generation helpers for new folder structure
// Pattern: users/{userID}/jobs/{jobID}/{type}/filename

func buildSceneClipKey(userID, jobID string, sceneNumber int) string {
	return fmt.Sprintf("users/%s/jobs/%s/clips/scene-%03d.mp4", userID, jobID, sceneNumber)
}

func buildSceneThumbnailKey(userID, jobID string, sceneNumber int) string {
	return fmt.Sprintf("users/%s/jobs/%s/thumbnails/scene-%03d.jpg", userID, jobID, sceneNumber)
}

func buildAudioKey(userID, jobID string) string {
	return fmt.Sprintf("users/%s/jobs/%s/audio/background-music.mp3", userID, jobID)
}

func buildFinalVideoKey(userID, jobID string) string {
	return fmt.Sprintf("users/%s/jobs/%s/final/video.mp4", userID, jobID)
}

func buildJobThumbnailKey(userID, jobID string) string {
	return fmt.Sprintf("users/%s/jobs/%s/thumbnails/job-thumbnail.jpg", userID, jobID)
}

// generateVideoAsync runs the entire video generation pipeline in a goroutine
func (h *GenerateHandler) generateVideoAsync(ctx context.Context, job *domain.Job, req GenerateRequest) {
	// Create job-specific context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, VideoGenerationTimeout)
	defer cancel()

	h.logger.Info("Starting async video generation",
		zap.String("job_id", job.JobID),
	)

	// STEP 1: Generate script with GPT-4o (happens in background now!)
	h.logger.Info("Generating script with GPT-4o", zap.String("job_id", job.JobID))
	job.Stage = "script_generating"
	if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
		h.logger.Error("Failed to update job stage",
			zap.String("job_id", job.JobID),
			zap.String("stage", "script_generating"),
			zap.Error(err),
		)
	}

	script, err := h.parserService.GenerateScript(jobCtx, service.ParseRequest{
		UserID:      job.UserID,
		Prompt:      req.Prompt,
		Duration:    req.Duration,
		AspectRatio: req.AspectRatio,
		StartImage:  req.StartImage,

		// Style reference image - will be analyzed and converted to text
		StyleReferenceImage: req.StyleReferenceImage,
	})
	if err != nil {
		h.logger.Error("Script generation failed", zap.String("job_id", job.JobID), zap.Error(err))
		h.jobRepo.MarkJobFailed(jobCtx, job.JobID, fmt.Sprintf("Script generation failed: %v", err))
		return
	}

	// Embed script in job record
	job.Title = script.Title
	job.Scenes = script.Scenes
	job.AudioSpec = script.AudioSpec
	job.ScriptMetadata = script.Metadata

	// Update job with embedded script
	job.Stage = "script_complete"
	if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
		h.logger.Error("Failed to update job stage",
			zap.String("job_id", job.JobID),
			zap.String("stage", "script_complete"),
			zap.Error(err),
		)
	}

	h.logger.Info("Script generated and embedded in job",
		zap.String("job_id", job.JobID),
		zap.String("title", script.Title),
		zap.Int("num_scenes", len(script.Scenes)),
		zap.String("audio_mood", script.AudioSpec.MusicMood),
		zap.String("audio_style", script.AudioSpec.MusicStyle),
	)

	// Log each scene for visibility
	for i, scene := range script.Scenes {
		h.logger.Info("Scene details",
			zap.String("job_id", job.JobID),
			zap.Int("scene_number", i+1),
			zap.Float64("start_time", scene.StartTime),
			zap.Float64("duration", scene.Duration),
			zap.String("shot_type", string(scene.ShotType)),
			zap.String("camera_angle", string(scene.CameraAngle)),
			zap.String("lighting", string(scene.Lighting)),
			zap.String("color_grade", string(scene.ColorGrade)),
			zap.String("mood", string(scene.Mood)),
			zap.String("generation_prompt", scene.GenerationPrompt),
		)
	}

	// STEP 2: Generate video clips sequentially
	var clipVideos []ClipVideo
	// Initialize lastFrameURL with user's start_image (for first scene)
	var lastFrameURL string = req.StartImage

	// Initialize arrays for accumulating scene data
	sceneVideoURLs := make([]string, 0, len(script.Scenes))
	var lastThumbnailURL string

	for i, scene := range script.Scenes {
		job.Stage = fmt.Sprintf("scene_%d_generating", i+1)
		job.ScenesCompleted = i // Number completed so far (i is 0-indexed)
		if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
			h.logger.Error("Failed to update job stage",
				zap.String("job_id", job.JobID),
				zap.String("stage", job.Stage),
				zap.Error(err),
			)
		}

		h.logger.Info("Generating scene",
			zap.String("job_id", job.JobID),
			zap.Int("scene", i+1),
			zap.Int("total", len(script.Scenes)),
		)

		// Image selection logic (style is now handled via text in generation_prompt):
		// 1. First scene: use user's start_image if provided
		// 2. Subsequent scenes: use last frame from previous clip for visual continuity
		scene.StartImageURL = lastFrameURL
		if i == 0 && lastFrameURL != "" {
			h.logger.Info("Using start image for first scene",
				zap.String("job_id", job.JobID),
				zap.String("start_image_url", lastFrameURL),
			)
		} else if lastFrameURL != "" {
			h.logger.Info("Using last frame for visual continuity",
				zap.String("job_id", job.JobID),
				zap.Int("scene", i+1),
			)
		}

		// Call Kling API (synchronous polling in this goroutine)
		clipResult, err := h.generateClip(jobCtx, job.UserID, job.JobID, scene, req.AspectRatio, i+1)
		if err != nil {
			h.logger.Error("Scene generation failed",
				zap.String("job_id", job.JobID),
				zap.Int("scene", i+1),
				zap.Error(err),
			)
			h.jobRepo.MarkJobFailed(jobCtx, job.JobID, fmt.Sprintf("Scene %d generation failed: %v", i+1, err))
			return
		}

		clipVideos = append(clipVideos, clipResult)
		lastFrameURL = clipResult.LastFrameURL

		// Accumulate scene data
		sceneVideoURLs = append(sceneVideoURLs, clipResult.VideoURL)
		lastThumbnailURL = clipResult.LastFrameURL

		// Extract and upload job thumbnail from first scene
		if i == 0 {
			jobThumbnail, err := h.extractJobThumbnail(jobCtx, job.UserID, job.JobID, clipResult.VideoURL)
			if err != nil {
				h.logger.Warn("Failed to extract job thumbnail, continuing without it",
					zap.String("job_id", job.JobID),
					zap.Error(err),
				)
				// Continue - this is not critical
			} else {
				// Use the job thumbnail (from first I-frame) instead of last frame URL
				lastThumbnailURL = jobThumbnail
				h.logger.Info("Job thumbnail set from first scene",
					zap.String("job_id", job.JobID),
					zap.String("thumbnail_url", jobThumbnail),
				)
			}
		}

		// Update job with accumulated data
		job.Stage = fmt.Sprintf("scene_%d_complete", i+1)
		job.ScenesCompleted = i + 1
		job.SceneVideoURLs = sceneVideoURLs
		job.ThumbnailURL = lastThumbnailURL

		if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
			h.logger.Error("Failed to update job progress",
				zap.String("job_id", job.JobID),
				zap.Int("scenes_completed", job.ScenesCompleted),
				zap.Error(err),
			)
		}

		h.logger.Info("Scene completed and job updated",
			zap.String("job_id", job.JobID),
			zap.Int("scene_number", i+1),
			zap.Int("total_scenes", len(script.Scenes)),
			zap.Int("scenes_completed", job.ScenesCompleted),
		)

		h.logger.Info("Scene completed",
			zap.String("job_id", job.JobID),
			zap.Int("scene", i+1),
			zap.String("video_url", clipResult.VideoURL),
		)
	}

	// STEP 3: Generate audio
	job.Stage = "audio_generating"
	if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
		h.logger.Error("Failed to update job stage",
			zap.String("job_id", job.JobID),
			zap.String("stage", "audio_generating"),
			zap.Error(err),
		)
	}
	h.logger.Info("Generating audio", zap.String("job_id", job.JobID))

	audioURL, err := h.generateAudio(jobCtx, job.UserID, job.JobID, script)
	if err != nil {
		h.logger.Error("Audio generation failed", zap.String("job_id", job.JobID), zap.Error(err))
		h.jobRepo.MarkJobFailed(jobCtx, job.JobID, fmt.Sprintf("Audio generation failed: %v", err))
		return
	}

	job.Stage = "audio_complete"
	job.AudioURL = audioURL

	if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
		h.logger.Error("Failed to update job with audio URL",
			zap.String("job_id", job.JobID),
			zap.Error(err),
		)
	}

	h.logger.Info("Audio generation complete",
		zap.String("job_id", job.JobID),
		zap.String("audio_url", audioURL),
	)

	// STEP 4: Compose final video
	job.Stage = "composing"
	if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
		h.logger.Error("Failed to update job stage",
			zap.String("job_id", job.JobID),
			zap.String("stage", "composing"),
			zap.Error(err),
		)
	}
	h.logger.Info("Composing final video", zap.String("job_id", job.JobID))

	finalVideoKey, err := h.composeVideo(jobCtx, job.UserID, job.JobID, clipVideos, audioURL)
	if err != nil {
		h.logger.Error("Video composition failed", zap.String("job_id", job.JobID), zap.Error(err))
		h.jobRepo.MarkJobFailed(jobCtx, job.JobID, fmt.Sprintf("Video composition failed: %v", err))
		return
	}

	// STEP 5: Mark job complete with video URL
	err = h.jobRepo.MarkJobComplete(jobCtx, job.JobID, finalVideoKey)
	if err != nil {
		h.logger.Error("Failed to mark job complete", zap.String("job_id", job.JobID), zap.Error(err))
		return
	}

	h.logger.Info("Video generation complete",
		zap.String("job_id", job.JobID),
		zap.String("video_key", finalVideoKey),
	)
}

// generateClip generates a single video clip using Kling AI
func (h *GenerateHandler) generateClip(
	ctx context.Context,
	userID string,
	jobID string,
	scene domain.Scene,
	aspectRatio string,
	clipNumber int,
) (ClipVideo, error) {
	h.logger.Info("Calling Kling adapter",
		zap.String("job_id", jobID),
		zap.Int("scene", scene.SceneNumber),
		zap.String("prompt", scene.GenerationPrompt),
	)

	// Call Kling adapter
	req := &adapters.VideoGenerationRequest{
		Prompt:        scene.GenerationPrompt,
		Duration:      int(scene.Duration),
		AspectRatio:   aspectRatio,
		StartImageURL: scene.StartImageURL,
	}

	result, err := h.klingAdapter.GenerateVideo(ctx, req)
	if err != nil {
		return ClipVideo{}, fmt.Errorf("kling API failed: %w", err)
	}

	// Poll until complete (max 10 minutes)
	maxAttempts := VideoGenerationMaxAttempts // 120 × 5s = 10 minutes
	pollInterval := PollInterval

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ClipVideo{}, ctx.Err()
		default:
		}

		if attempt > 0 {
			time.Sleep(pollInterval)
			result, err = h.klingAdapter.GetStatus(ctx, result.PredictionID)
			if err != nil {
				h.logger.Warn("Kling polling failed, retrying", zap.Error(err))
				continue
			}
		}

		if result.Status == "succeeded" || result.Status == "completed" {
			// Download video, extract last frame, upload to S3
			clipURL, lastFrameURL, err := h.processVideo(ctx, userID, jobID, clipNumber, result.VideoURL)
			if err != nil {
				return ClipVideo{}, fmt.Errorf("video processing failed: %w", err)
			}

			return ClipVideo{
				VideoURL:     clipURL,
				LastFrameURL: lastFrameURL,
				Duration:     scene.Duration,
			}, nil
		}

		if result.Status == "failed" || result.Status == "canceled" {
			return ClipVideo{}, fmt.Errorf("kling generation failed: %s", result.Error)
		}

		// Log only every 12th attempt (every minute instead of every 5 seconds)
		if attempt%12 == 0 {
			h.logger.Debug("Kling still processing",
				zap.String("job_id", jobID),
				zap.Int("attempt", attempt),
				zap.String("status", result.Status),
			)
		}
	}

	return ClipVideo{}, fmt.Errorf("clip generation timed out after %d attempts", maxAttempts)
}

// processVideo downloads video from Replicate, extracts last frame, uploads both to S3
func (h *GenerateHandler) processVideo(
	ctx context.Context,
	userID string,
	jobID string,
	clipNumber int,
	videoURL string,
) (string, string, error) {
	// Create temp directory
	tmpDir := filepath.Join("/tmp", jobID, fmt.Sprintf("clip-%d", clipNumber))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download video from Replicate URL
	h.logger.Info("Downloading video from Replicate",
		zap.String("job_id", jobID),
		zap.Int("clip", clipNumber),
		zap.String("url", videoURL),
	)
	videoPath := filepath.Join(tmpDir, "video.mp4")
	if err := h.downloadFile(ctx, videoURL, videoPath); err != nil {
		return "", "", fmt.Errorf("failed to download video: %w", err)
	}

	// Extract last frame using ffmpeg
	h.logger.Info("Extracting last frame with ffmpeg",
		zap.String("job_id", jobID),
		zap.Int("clip", clipNumber),
	)
	lastFramePath := filepath.Join(tmpDir, "last_frame.jpg")
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-sseof", "-1",
		"-i", videoPath,
		"-update", "1",
		"-q:v", "2",
		"-y", lastFramePath,
	)
	if err := cmd.Run(); err != nil {
		h.logger.Warn("Failed to extract last frame, continuing without it",
			zap.String("job_id", jobID),
			zap.Int("clip", clipNumber),
			zap.Error(err),
		)
		lastFramePath = "" // Continue without last frame
	} else {
		h.logger.Info("Last frame extracted successfully",
			zap.String("job_id", jobID),
			zap.Int("clip", clipNumber),
		)
	}

	// Upload video to S3
	h.logger.Info("Uploading video to S3",
		zap.String("job_id", jobID),
		zap.Int("clip", clipNumber),
	)
	videoS3Key := buildSceneClipKey(userID, jobID, clipNumber)
	videoS3URL, err := h.s3Service.UploadFile(ctx, h.assetsBucket, videoS3Key, videoPath, "video/mp4")
	if err != nil {
		return "", "", fmt.Errorf("failed to upload video to S3: %w", err)
	}

	// Upload last frame to S3 (if extracted)
	var lastFrameS3URL string
	if lastFramePath != "" {
		lastFrameS3Key := buildSceneThumbnailKey(userID, jobID, clipNumber)
		_, err = h.s3Service.UploadFile(ctx, h.assetsBucket, lastFrameS3Key, lastFramePath, "image/jpeg")
		if err != nil {
			h.logger.Warn("Failed to upload last frame, continuing",
				zap.String("job_id", jobID),
				zap.Error(err),
			)
			lastFrameS3URL = "" // Continue without last frame URL
		} else {
			// Generate presigned URL for Kling API access (valid for 1 hour)
			lastFrameS3URL, err = h.s3Service.GetPresignedURL(ctx, lastFrameS3Key, 1*time.Hour)
			if err != nil {
				h.logger.Warn("Failed to generate presigned URL for last frame, continuing",
					zap.String("job_id", jobID),
					zap.Error(err),
				)
				lastFrameS3URL = "" // Continue without last frame URL
			}
		}
	}

	h.logger.Info("Video processed and uploaded",
		zap.String("job_id", jobID),
		zap.Int("clip", clipNumber),
		zap.String("s3_url", videoS3URL),
	)

	return videoS3URL, lastFrameS3URL, nil
}

// extractJobThumbnail extracts a middle frame from the first scene video and uploads as job thumbnail
func (h *GenerateHandler) extractJobThumbnail(
	ctx context.Context,
	userID string,
	jobID string,
	videoURL string,
) (string, error) {
	h.logger.Info("Extracting job thumbnail from first scene",
		zap.String("job_id", jobID),
		zap.String("video_url", videoURL),
	)

	// Create temp directory
	tmpDir := filepath.Join("/tmp", jobID, "thumbnail")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download video
	videoPath := filepath.Join(tmpDir, "video.mp4")
	if err := h.downloadFile(ctx, videoURL, videoPath); err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	// Extract middle frame (at 50% of video duration for best representation)
	thumbnailPath := filepath.Join(tmpDir, "thumbnail.jpg")
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", videoPath,
		"-vf", "select='eq(pict_type\\,I)',scale=1280:-1", // Extract I-frame, scale to 1280px width
		"-frames:v", "1",
		"-q:v", "2", // High quality
		"-y", thumbnailPath,
	)

	if err := cmd.Run(); err != nil {
		h.logger.Warn("Failed to extract thumbnail, trying simpler method",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		// Fallback: extract frame at 1 second
		cmd = exec.CommandContext(ctx, "ffmpeg",
			"-ss", "1",
			"-i", videoPath,
			"-frames:v", "1",
			"-q:v", "2",
			"-y", thumbnailPath,
		)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to extract thumbnail: %w", err)
		}
	}

	// Upload to S3
	thumbnailS3Key := buildJobThumbnailKey(userID, jobID)
	thumbnailURL, err := h.s3Service.UploadFile(ctx, h.assetsBucket, thumbnailS3Key, thumbnailPath, "image/jpeg")
	if err != nil {
		return "", fmt.Errorf("failed to upload thumbnail to S3: %w", err)
	}

	h.logger.Info("Job thumbnail extracted and uploaded successfully",
		zap.String("job_id", jobID),
		zap.String("thumbnail_url", thumbnailURL),
	)

	return thumbnailURL, nil
}

// generateAudio generates background music using Minimax
func (h *GenerateHandler) generateAudio(
	ctx context.Context,
	userID string,
	jobID string,
	script *domain.Script,
) (string, error) {
	h.logger.Info("Calling Minimax adapter", zap.String("job_id", jobID))

	req := &adapters.MusicGenerationRequest{
		Prompt:     script.Title,
		Duration:   script.TotalDuration,
		MusicMood:  script.AudioSpec.MusicMood,
		MusicStyle: script.AudioSpec.MusicStyle,
	}

	result, err := h.minimaxAdapter.GenerateMusic(ctx, req)
	if err != nil {
		return "", fmt.Errorf("minimax API failed: %w", err)
	}

	// Poll until complete (max 5 minutes)
	maxAttempts := AudioGenerationMaxAttempts // 60 × 5s = 5 minutes
	pollInterval := PollInterval

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if attempt > 0 {
			time.Sleep(pollInterval)
			result, err = h.minimaxAdapter.GetStatus(ctx, result.PredictionID)
			if err != nil {
				h.logger.Warn("Minimax polling failed, retrying", zap.Error(err))
				continue
			}
		}

		if result.Status == "succeeded" || result.Status == "completed" {
			// Download and upload to S3
			audioS3URL, err := h.processAudio(ctx, userID, jobID, result.AudioURL)
			if err != nil {
				return "", fmt.Errorf("audio processing failed: %w", err)
			}
			return audioS3URL, nil
		}

		if result.Status == "failed" || result.Status == "canceled" {
			return "", fmt.Errorf("minimax generation failed: %s", result.Error)
		}

		// Log only every 12th attempt (every minute instead of every 5 seconds)
		if attempt%12 == 0 {
			h.logger.Debug("Minimax still processing",
				zap.String("job_id", jobID),
				zap.Int("attempt", attempt),
			)
		}
	}

	return "", fmt.Errorf("audio generation timed out")
}

// processAudio downloads audio from Replicate, uploads to S3
func (h *GenerateHandler) processAudio(ctx context.Context, userID string, jobID string, audioURL string) (string, error) {
	tmpDir := filepath.Join("/tmp", jobID, "audio")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download audio
	h.logger.Info("Downloading audio from Replicate",
		zap.String("job_id", jobID),
		zap.String("url", audioURL),
	)
	audioPath := filepath.Join(tmpDir, "music.mp3")
	if err := h.downloadFile(ctx, audioURL, audioPath); err != nil {
		return "", fmt.Errorf("failed to download audio: %w", err)
	}

	// Upload to S3
	h.logger.Info("Uploading audio to S3",
		zap.String("job_id", jobID),
	)
	audioS3Key := buildAudioKey(userID, jobID)
	audioS3URL, err := h.s3Service.UploadFile(ctx, h.assetsBucket, audioS3Key, audioPath, "audio/mpeg")
	if err != nil {
		return "", fmt.Errorf("failed to upload audio to S3: %w", err)
	}

	h.logger.Info("Audio processed and uploaded", zap.String("job_id", jobID), zap.String("s3_url", audioS3URL))
	return audioS3URL, nil
}

// composeVideo stitches clips together with audio using ffmpeg
func (h *GenerateHandler) composeVideo(
	ctx context.Context,
	userID string,
	jobID string,
	clips []ClipVideo,
	audioURL string,
) (string, error) {
	h.logger.Info("Composing video with ffmpeg",
		zap.String("job_id", jobID),
		zap.Int("num_clips", len(clips)),
	)

	tmpDir := filepath.Join("/tmp", jobID, "composition")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download all clips from S3
	h.logger.Info("Downloading clips from S3 for composition",
		zap.String("job_id", jobID),
		zap.Int("num_clips", len(clips)),
	)
	var clipPaths []string
	for i, clip := range clips {
		clipPath := filepath.Join(tmpDir, fmt.Sprintf("clip-%d.mp4", i+1))
		if err := h.s3Service.DownloadFile(ctx, h.assetsBucket, extractS3Key(clip.VideoURL), clipPath); err != nil {
			return "", fmt.Errorf("failed to download clip %d: %w", i+1, err)
		}
		clipPaths = append(clipPaths, clipPath)
	}

	// Download audio from S3
	h.logger.Info("Downloading audio from S3 for composition",
		zap.String("job_id", jobID),
	)
	audioPath := filepath.Join(tmpDir, "music.mp3")
	if err := h.s3Service.DownloadFile(ctx, h.assetsBucket, extractS3Key(audioURL), audioPath); err != nil {
		return "", fmt.Errorf("failed to download audio: %w", err)
	}

	// Create concat file for ffmpeg
	concatFile := filepath.Join(tmpDir, "concat.txt")
	f, err := os.Create(concatFile)
	if err != nil {
		return "", fmt.Errorf("failed to create concat file: %w", err)
	}
	for _, path := range clipPaths {
		fmt.Fprintf(f, "file '%s'\n", path)
	}
	f.Close()

	// Concatenate clips
	h.logger.Info("Concatenating video clips with ffmpeg",
		zap.String("job_id", jobID),
		zap.Int("num_clips", len(clipPaths)),
	)
	videoNoAudio := filepath.Join(tmpDir, "video_no_audio.mp4")
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c", "copy",
		"-y", videoNoAudio,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		h.logger.Error("ffmpeg concat failed",
			zap.String("job_id", jobID),
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return "", fmt.Errorf("ffmpeg concat failed: %w", err)
	}

	// Add audio
	h.logger.Info("Merging audio track with ffmpeg",
		zap.String("job_id", jobID),
	)
	finalVideo := filepath.Join(tmpDir, "final.mp4")
	cmd = exec.CommandContext(ctx, "ffmpeg",
		"-i", videoNoAudio,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-shortest",
		"-y", finalVideo,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		h.logger.Error("ffmpeg audio merge failed",
			zap.String("job_id", jobID),
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return "", fmt.Errorf("ffmpeg audio merge failed: %w", err)
	}

	// Upload final video to S3
	h.logger.Info("Uploading final video to S3",
		zap.String("job_id", jobID),
	)
	finalS3Key := buildFinalVideoKey(userID, jobID)
	finalURL, err := h.s3Service.UploadFile(ctx, h.assetsBucket, finalS3Key, finalVideo, "video/mp4")
	if err != nil {
		return "", fmt.Errorf("failed to upload final video: %w", err)
	}

	h.logger.Info("Video composition complete",
		zap.String("job_id", jobID),
		zap.String("final_url", finalURL),
	)

	return finalS3Key, nil
}

// downloadFile downloads a file from URL to local path
func (h *GenerateHandler) downloadFile(ctx context.Context, url string, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractS3Key extracts the S3 key from an S3 URL
func extractS3Key(s3URL string) string {
	// Strip query parameters first (handles presigned URLs that may have been stored)
	if idx := strings.Index(s3URL, "?"); idx != -1 {
		s3URL = s3URL[:idx]
	}

	// Extract key from URL like https://bucket.s3.amazonaws.com/key or https://s3.amazonaws.com/bucket/key
	parts := strings.Split(s3URL, "/")
	if len(parts) < 4 {
		return s3URL
	}

	// Find where the key starts (after bucket name)
	for i, part := range parts {
		if strings.HasSuffix(part, ".amazonaws.com") {
			if i+1 < len(parts) {
				return strings.Join(parts[i+1:], "/")
			}
		}
	}

	// Fallback: return everything after the first 3 slashes
	return strings.Join(parts[3:], "/")
}
