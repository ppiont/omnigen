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

// generateVideoAsync runs the entire video generation pipeline in a goroutine
func (h *GenerateHandler) generateVideoAsync(ctx context.Context, job *domain.Job, req GenerateRequest) {
	// Create job-specific context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, VideoGenerationTimeout)
	defer cancel()

	h.logger.Info("Starting async video generation",
		zap.String("job_id", job.JobID),
	)

	// Helper to update job stage and metadata
	updateStage := func(stage string, metadata map[string]interface{}) {
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		err := h.jobRepo.UpdateJobStageWithMetadata(jobCtx, job.JobID, stage, metadata)
		if err != nil {
			h.logger.Error("Failed to update job stage",
				zap.String("job_id", job.JobID),
				zap.String("stage", stage),
				zap.Error(err),
			)
		}
	}

	// STEP 1: Generate script with GPT-4o (happens in background now!)
	h.logger.Info("Generating script with GPT-4o", zap.String("job_id", job.JobID))
	updateStage("script_generating", nil)

	script, err := h.parserService.GenerateScript(jobCtx, service.ParseRequest{
		UserID:      job.UserID,
		Prompt:      req.Prompt,
		Duration:    req.Duration,
		AspectRatio: req.AspectRatio,
		StartImage:  req.StartImage,

		// Style reference image - will be analyzed and converted to text
		StyleReferenceImage: req.StyleReferenceImage,

		// Enhanced prompt options (Phase 1)
		Style:             req.Style,
		Tone:              req.Tone,
		Tempo:             req.Tempo,
		Platform:          req.Platform,
		Audience:          req.Audience,
		Goal:              req.Goal,
		CallToAction:      req.CallToAction,
		ProCinematography: req.ProCinematography,
		CreativeBoost:     req.CreativeBoost,
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
	updateStage("script_complete", map[string]interface{}{
		"num_scenes": len(script.Scenes),
		"audio_mood": script.AudioSpec.MusicMood,
	})

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

	for i, scene := range script.Scenes {
		updateStage(fmt.Sprintf("scene_%d_generating", i+1), map[string]interface{}{
			"current_scene": i + 1,
			"total_scenes":  len(script.Scenes),
		})

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
		clipResult, err := h.generateClip(jobCtx, job.JobID, scene, req.AspectRatio, i+1)
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

		// Update with scene completion + thumbnail
		updateStage(fmt.Sprintf("scene_%d_complete", i+1), map[string]interface{}{
			"thumbnail_url":   clipResult.LastFrameURL,
			"scenes_complete": i + 1,
			"scenes_total":    len(script.Scenes),
			"scene_video_url": clipResult.VideoURL,
		})

		h.logger.Info("Scene completed",
			zap.String("job_id", job.JobID),
			zap.Int("scene", i+1),
			zap.String("video_url", clipResult.VideoURL),
		)
	}

	// STEP 3: Generate audio
	updateStage("audio_generating", nil)
	h.logger.Info("Generating audio", zap.String("job_id", job.JobID))

	audioURL, err := h.generateAudio(jobCtx, job.JobID, script)
	if err != nil {
		h.logger.Error("Audio generation failed", zap.String("job_id", job.JobID), zap.Error(err))
		h.jobRepo.MarkJobFailed(jobCtx, job.JobID, fmt.Sprintf("Audio generation failed: %v", err))
		return
	}

	updateStage("audio_complete", map[string]interface{}{
		"audio_url": audioURL,
	})

	h.logger.Info("Audio generated", zap.String("job_id", job.JobID), zap.String("audio_url", audioURL))

	// STEP 4: Compose final video
	updateStage("composing", nil)
	h.logger.Info("Composing final video", zap.String("job_id", job.JobID))

	finalVideoKey, err := h.composeVideo(jobCtx, job.JobID, clipVideos, audioURL)
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
			clipURL, lastFrameURL, err := h.processVideo(ctx, jobID, clipNumber, result.VideoURL)
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
	videoS3Key := fmt.Sprintf("videos/%s/clip-%d.mp4", jobID, clipNumber)
	videoS3URL, err := h.s3Service.UploadFile(ctx, h.assetsBucket, videoS3Key, videoPath, "video/mp4")
	if err != nil {
		return "", "", fmt.Errorf("failed to upload video to S3: %w", err)
	}

	// Upload last frame to S3 (if extracted)
	var lastFrameS3URL string
	if lastFramePath != "" {
		lastFrameS3Key := fmt.Sprintf("videos/%s/clip-%d-last-frame.jpg", jobID, clipNumber)
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

// generateAudio generates background music using Minimax
func (h *GenerateHandler) generateAudio(
	ctx context.Context,
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
			audioS3URL, err := h.processAudio(ctx, jobID, result.AudioURL)
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
func (h *GenerateHandler) processAudio(ctx context.Context, jobID string, audioURL string) (string, error) {
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
	audioS3Key := fmt.Sprintf("videos/%s/music.mp3", jobID)
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
	finalS3Key := fmt.Sprintf("videos/%s/final.mp4", jobID)
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
