package handlers

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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

// S3 Key Generation Helpers
//
// These helpers centralize S3 key construction for all job assets.
// Every key follows the pattern: users/{userID}/jobs/{jobID}/{type}/{filename}
//
// Folder structure:
//
//   users/{userID}/jobs/{jobID}/
//     ├── clips/
//     │   ├── scene-001.mp4          (buildSceneClipKey)
//     │   ├── scene-002.mp4
//     │   └── scene-NNN.mp4
//     ├── thumbnails/
//     │   ├── scene-001.jpg          (buildSceneThumbnailKey)
//     │   ├── scene-002.jpg
//     │   └── job-thumbnail.jpg      (buildJobThumbnailKey)
//     ├── audio/
//     │   ├── background-music.mp3   (buildAudioKey)
//     │   └── narrator-voiceover.mp3 (buildNarratorAudioKey)
//     └── final/
//         └── video.mp4               (buildFinalVideoKey)
//
// Usage notes:
//   - Clips: Raw scene videos generated per scene (no audio)
//   - Thumbnails: Preview frames used for UI cards and continuity
//   - Audio: Separate tracks (music via Minimax, narrator via TTS)
//   - Final: Composited video without audio tracks, ready for playback

func buildSceneClipKey(userID, jobID string, sceneNumber int) string {
	return fmt.Sprintf("users/%s/jobs/%s/clips/scene-%03d.mp4", userID, jobID, sceneNumber)
}

func buildSceneThumbnailKey(userID, jobID string, sceneNumber int) string {
	return fmt.Sprintf("users/%s/jobs/%s/thumbnails/scene-%03d.jpg", userID, jobID, sceneNumber)
}

func buildAudioKey(userID, jobID string) string {
	return fmt.Sprintf("users/%s/jobs/%s/audio/background-music.mp3", userID, jobID)
}

func buildNarratorAudioKey(userID, jobID string) string {
	return fmt.Sprintf("users/%s/jobs/%s/audio/narrator-voiceover.mp3", userID, jobID)
}

func buildFinalVideoKey(userID, jobID string) string {
	return fmt.Sprintf("users/%s/jobs/%s/final/video.mp4", userID, jobID)
}

func buildJobThumbnailKey(userID, jobID string) string {
	return fmt.Sprintf("users/%s/jobs/%s/thumbnails/job-thumbnail.jpg", userID, jobID)
}

const (
	scriptFailureMessage      = "Script generation failed. Please check your prompt and try again."
	narratorFailureMessage    = "Voiceover generation failed. Please try again later."
	sceneFailureMessageFormat = "Video generation failed at scene %d. Please try again."
	audioFailureMessage       = "Background music generation failed. Please try again."
	compositionFailureMessage = "Video composition failed. Please try again."
)

func (h *GenerateHandler) failJob(
	ctx context.Context,
	job *domain.Job,
	userMessage string,
	internalErr error,
	fields ...zap.Field,
) {
	logFields := []zap.Field{
		zap.String("job_id", job.JobID),
		zap.String("user_message", userMessage),
	}
	if internalErr != nil {
		logFields = append(logFields, zap.Error(internalErr))
	}
	if len(fields) > 0 {
		logFields = append(logFields, fields...)
	}

	h.logger.Error("Job failed", logFields...)
	h.cleanupJobAssets(job.UserID, job.JobID)

	// Build detailed error message for user
	errorMessage := userMessage
	if internalErr != nil {
		// Extract meaningful error details
		errStr := internalErr.Error()

		// Add technical details in a user-friendly way
		// Check for common error patterns and provide helpful context
		if strings.Contains(errStr, "Payment required") || strings.Contains(errStr, "status 402") || strings.Contains(errStr, "status 402") {
			// HTTP 402 - Payment Required (Replicate credits/billing issue)
			errorMessage = "Script generation failed due to insufficient Replicate API credits. Please check your Replicate account balance and billing settings."
		} else if strings.Contains(errStr, "API error") || (strings.Contains(errStr, "status") && !strings.Contains(errStr, "exit status")) {
			// API errors - include status code if available (but not ffmpeg/process exit codes)
			// For 422 errors, try to extract more details from the response
			if strings.Contains(errStr, "422") {
				// Try to extract the actual error message from Replicate
				if strings.Contains(errStr, "Response:") {
					// Extract the response body which should contain the actual error
					parts := strings.Split(errStr, "Response:")
					if len(parts) > 1 {
						responseBody := strings.TrimSpace(parts[1])
						// Limit response body length
						if len(responseBody) > 200 {
							responseBody = responseBody[:200] + "..."
						}
						errorMessage = fmt.Sprintf("%s (API Error: HTTP 422 - %s)", userMessage, responseBody)
					} else {
						errorMessage = fmt.Sprintf("%s (API Error: %s)", userMessage, extractAPIError(errStr))
					}
				} else {
					errorMessage = fmt.Sprintf("%s (API Error: %s)", userMessage, extractAPIError(errStr))
				}
			} else {
				errorMessage = fmt.Sprintf("%s (API Error: %s)", userMessage, extractAPIError(errStr))
			}
		} else if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline") {
			errorMessage = fmt.Sprintf("%s (Request timed out. The service may be busy. Please try again.)", userMessage)
		} else if strings.Contains(errStr, "authentication") || strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "401") {
			errorMessage = fmt.Sprintf("%s (Authentication failed. Please check API configuration.)", userMessage)
		} else if strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "429") {
			errorMessage = fmt.Sprintf("%s (Rate limit exceeded. Please wait a moment and try again.)", userMessage)
		} else {
			// For other errors, include a sanitized version of the error
			// Truncate very long error messages
			if len(errStr) > 200 {
				errStr = errStr[:200] + "..."
			}
			errorMessage = fmt.Sprintf("%s (Error: %s)", userMessage, errStr)
		}
	}

	if err := h.jobRepo.MarkJobFailed(ctx, job.JobID, errorMessage); err != nil {
		h.logger.Error("Failed to mark job failed",
			zap.String("job_id", job.JobID),
			zap.Error(err),
		)
	}
}

// extractAPIError extracts meaningful information from API error messages
func extractAPIError(errStr string) string {
	// Try to extract status code
	if strings.Contains(errStr, "status") {
		// Look for patterns like "status 401" or "status 500"
		parts := strings.Split(errStr, "status")
		if len(parts) > 1 {
			statusPart := strings.TrimSpace(parts[1])
			// Extract just the status code
			statusCode := ""
			for _, char := range statusPart {
				if char >= '0' && char <= '9' {
					statusCode += string(char)
				} else if statusCode != "" {
					break
				}
			}
			if statusCode != "" {
				return fmt.Sprintf("HTTP %s", statusCode)
			}
		}
	}

	// Fallback: return first 100 chars
	if len(errStr) > 100 {
		return errStr[:100] + "..."
	}
	return errStr
}

func (h *GenerateHandler) cleanupJobAssets(userID, jobID string) {
	if h.s3Service == nil || h.assetsBucket == "" {
		return
	}

	cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prefix := fmt.Sprintf("users/%s/jobs/%s/", userID, jobID)
	if err := h.s3Service.DeletePrefix(cleanupCtx, h.assetsBucket, prefix); err != nil {
		h.logger.Warn("Failed to cleanup S3 assets after job failure",
			zap.String("job_id", jobID),
			zap.String("prefix", prefix),
			zap.Error(err),
		)
	}
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

		// Pharmaceutical ad configuration
		Voice:       job.Voice,
		SideEffects: job.SideEffects,

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
		h.logger.Error("Script generation failed with error",
			zap.String("job_id", job.JobID),
			zap.String("stage", "script_generating"),
			zap.Error(err),
			zap.String("error_type", fmt.Sprintf("%T", err)),
			zap.String("error_string", err.Error()),
		)
		h.failJob(jobCtx, job, scriptFailureMessage, err, zap.String("stage", "script_generating"))
		return
	}

	// Embed script in job record
	job.Title = script.Title
	job.Scenes = script.Scenes
	job.AudioSpec = script.AudioSpec
	job.ScriptMetadata = script.Metadata

	// ALWAYS use the user's original side effects text for FDA compliance
	// GPT-4o should NOT generate or modify side effects - this is legally required verbatim text
	job.SideEffectsText = job.SideEffects
	if job.SideEffectsText != "" {
		// Default to 80% of duration for side effects start time
		job.SideEffectsStartTime = float64(job.Duration) * 0.8
		h.logger.Info("Using user-provided side effects text for FDA compliance",
			zap.String("job_id", job.JobID),
			zap.Int("text_length", len(job.SideEffectsText)),
			zap.Float64("side_effects_start_time", job.SideEffectsStartTime),
		)
	}

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

	// STEP 2: Generate narrator voiceover (if configured)
	if job.Voice != "" && job.AudioSpec.NarratorScript != "" {
		job.Stage = "narrator_generating"
		if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
			h.logger.Error("Failed to update job stage",
				zap.String("job_id", job.JobID),
				zap.String("stage", "narrator_generating"),
				zap.Error(err),
			)
		}

		narratorAudioURL, err := h.generateNarratorVoiceover(
			jobCtx,
			job.UserID,
			job.JobID,
			job.Voice,
			job.AudioSpec.NarratorScript,
			job.AudioSpec.SideEffectsStartTime,
			job.Duration,
		)
		if err != nil {
			h.failJob(jobCtx, job, narratorFailureMessage, err, zap.String("stage", "narrator_generating"))
			return
		}

		job.Stage = "narrator_complete"
		job.NarratorAudioURL = narratorAudioURL

		if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
			h.logger.Error("Failed to update job with narrator audio URL",
				zap.String("job_id", job.JobID),
				zap.Error(err),
			)
		}

		h.logger.Info("Narrator voiceover generation complete",
			zap.String("job_id", job.JobID),
			zap.String("narrator_audio_url", narratorAudioURL),
		)
	} else {
		if job.Voice == "" {
			h.logger.Info("Skipping narrator voiceover generation (voice not configured)",
				zap.String("job_id", job.JobID),
			)
		} else {
			h.logger.Warn("Skipping narrator voiceover generation (narrator script missing)",
				zap.String("job_id", job.JobID),
				zap.Int("script_length", len(job.AudioSpec.NarratorScript)),
			)
		}
	}

	// STEP 3: Start audio generation in parallel with video generation
	type audioResult struct {
		audioURL string
		err      error
	}

	audioChan := make(chan audioResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("Panic in audio generation",
					zap.String("job_id", job.JobID),
					zap.Any("panic", r),
				)
				audioChan <- audioResult{err: fmt.Errorf("audio generation panic: %v", r)}
			}
		}()

		job.Stage = "audio_generating"
		if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
			h.logger.Error("Failed to update job stage",
				zap.String("job_id", job.JobID),
				zap.String("stage", "audio_generating"),
				zap.Error(err),
			)
		}
		h.logger.Info("Generating audio (parallel with video)", zap.String("job_id", job.JobID))

		audioURL, err := h.generateAudio(jobCtx, job.UserID, job.JobID, script)
		audioChan <- audioResult{audioURL: audioURL, err: err}
	}()

	// STEP 4: Generate video clips sequentially
	var clipVideos []ClipVideo
	// Start with empty lastFrameURL so the first scene is pure AI generation
	var lastFrameURL string

	// Initialize arrays for accumulating scene data
	sceneVideoURLs := make([]string, 0, len(script.Scenes))
	var jobThumbnailURL string // Raw S3 URL for the job thumbnail (not presigned)

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

		// Image selection logic for pharmaceutical ads:
		// 1. Scenes 1..N-1 use the previous clip's last frame for continuity.
		// 2. Last scene (N) uses the product image provided by the user.
		if i == len(script.Scenes)-1 && strings.TrimSpace(req.StartImage) != "" {
			// Extract S3 key from the product image URL
			s3Key := extractS3Key(req.StartImage)

			// Generate presigned URL for video API access (valid for 1 hour)
			presignedURL, err := h.s3Service.GetPresignedURL(jobCtx, s3Key, 1*time.Hour)
			if err != nil {
				h.logger.Error("Failed to generate presigned URL for product image",
					zap.String("job_id", job.JobID),
					zap.String("s3_key", s3Key),
					zap.String("original_url", req.StartImage),
					zap.Error(err),
				)
				// Fall back to direct URL if presigning fails (shouldn't happen, but be safe)
				scene.StartImageURL = req.StartImage
			} else {
				scene.StartImageURL = presignedURL
				h.logger.Info("Using product image for last scene (side effects segment)",
					zap.String("job_id", job.JobID),
					zap.Int("scene", i+1),
					zap.String("product_image_url", presignedURL),
				)
			}
		} else {
			scene.StartImageURL = lastFrameURL
			if lastFrameURL != "" {
				h.logger.Info("Using last frame for visual continuity",
					zap.String("job_id", job.JobID),
					zap.Int("scene", i+1),
				)
			} else if i != len(script.Scenes)-1 {
				h.logger.Info("No continuity frame available, generating scene without start image",
					zap.String("job_id", job.JobID),
					zap.Int("scene", i+1),
				)
			} else {
				h.logger.Warn("Product image missing for last scene; falling back to continuity frame",
					zap.String("job_id", job.JobID),
					zap.Int("scene", i+1),
				)
			}
		}

		// Call Veo API (synchronous polling in this goroutine)
		clipResult, err := h.generateClip(jobCtx, job.UserID, job.JobID, scene, req.AspectRatio, i+1)
		if err != nil {
			h.failJob(jobCtx, job, fmt.Sprintf(sceneFailureMessageFormat, i+1), err,
				zap.String("stage", fmt.Sprintf("scene_%d_generating", i+1)),
				zap.Int("scene", i+1),
			)
			return
		}

		clipVideos = append(clipVideos, clipResult)
		lastFrameURL = clipResult.LastFrameURL

		// Accumulate scene data
		sceneVideoURLs = append(sceneVideoURLs, clipResult.VideoURL)

		// Extract and upload job thumbnail from first scene only
		// Note: clipResult.LastFrameURL is presigned (for Veo API continuity) and should NOT be stored in DB
		if i == 0 {
			jobThumbnail, err := h.extractJobThumbnail(jobCtx, job.UserID, job.JobID, clipResult.VideoURL)
			if err != nil {
				h.logger.Warn("Failed to extract job thumbnail, continuing without it",
					zap.String("job_id", job.JobID),
					zap.Error(err),
				)
				// Continue - this is not critical
			} else {
				// Store raw S3 URL (not presigned) - will be presigned when served via API
				jobThumbnailURL = jobThumbnail
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
		job.ThumbnailURL = jobThumbnailURL

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

	// STEP 5: Wait for audio generation to complete
	h.logger.Info("Waiting for audio generation to complete", zap.String("job_id", job.JobID))
	audioRes := <-audioChan
	if audioRes.err != nil {
		h.failJob(jobCtx, job, audioFailureMessage, audioRes.err, zap.String("stage", "audio_generating"))
		return
	}

	job.Stage = "audio_complete"
	job.AudioURL = audioRes.audioURL

	if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
		h.logger.Error("Failed to update job with audio URL",
			zap.String("job_id", job.JobID),
			zap.Error(err),
		)
	}

	h.logger.Info("Audio generation complete",
		zap.String("job_id", job.JobID),
		zap.String("audio_url", audioRes.audioURL),
	)

	// STEP 6: Compose final video
	job.Stage = "composing"
	if err := h.jobRepo.UpdateJob(jobCtx, job); err != nil {
		h.logger.Error("Failed to update job stage",
			zap.String("job_id", job.JobID),
			zap.String("stage", "composing"),
			zap.Error(err),
		)
	}
	h.logger.Info("Composing final video (video track only)", zap.String("job_id", job.JobID))

	finalVideoKey, err := h.composeVideo(
		jobCtx,
		job.UserID,
		job.JobID,
		clipVideos,
		job.SideEffectsText,
		job.SideEffectsStartTime,
	)
	if err != nil {
		h.failJob(jobCtx, job, compositionFailureMessage, err, zap.String("stage", "composing"))
		return
	}

	// STEP 6: Mark job complete with video URL
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

// generateClip generates a single video clip using Veo 3.1
func (h *GenerateHandler) generateClip(
	ctx context.Context,
	userID string,
	jobID string,
	scene domain.Scene,
	aspectRatio string,
	clipNumber int,
) (ClipVideo, error) {
	h.logger.Info("Calling Veo adapter",
		zap.String("job_id", jobID),
		zap.Int("scene", scene.SceneNumber),
		zap.String("prompt", scene.GenerationPrompt),
	)

	// Call Veo adapter
	req := &adapters.VideoGenerationRequest{
		Prompt:        scene.GenerationPrompt,
		Duration:      int(scene.Duration),
		AspectRatio:   aspectRatio,
		StartImageURL: scene.StartImageURL,
	}

	result, err := h.veoAdapter.GenerateVideo(ctx, req)
	if err != nil {
		return ClipVideo{}, fmt.Errorf("veo API failed: %w", err)
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
			result, err = h.veoAdapter.GetStatus(ctx, result.PredictionID)
			if err != nil {
				h.logger.Warn("Veo polling failed, retrying", zap.Error(err))
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
			// Provide better error message if error is empty
			errorMsg := result.Error
			if errorMsg == "" {
				errorMsg = "Unknown error - Veo returned failed status without error details"
			}
			h.logger.Error("Veo generation failed",
				zap.String("job_id", jobID),
				zap.Int("scene", scene.SceneNumber),
				zap.String("prediction_id", result.PredictionID),
				zap.String("error", errorMsg),
			)
			return ClipVideo{}, fmt.Errorf("veo generation failed: %s", errorMsg)
		}

		// Log only every 12th attempt (every minute instead of every 5 seconds)
		if attempt%12 == 0 {
			h.logger.Info("Veo still processing",
				zap.String("job_id", jobID),
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", maxAttempts),
				zap.String("status", result.Status),
				zap.String("prediction_id", result.PredictionID),
			)
		}

		// Log every 60th attempt (every 5 minutes) with more detail
		if attempt > 0 && attempt%60 == 0 {
			h.logger.Warn("Veo generation taking longer than expected",
				zap.String("job_id", jobID),
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", maxAttempts),
				zap.String("status", result.Status),
				zap.String("prediction_id", result.PredictionID),
				zap.String("error", result.Error),
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
			// Generate presigned URL for Veo API access (valid for 1 hour)
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

	// Extract the very first frame of the video for the thumbnail
	thumbnailPath := filepath.Join(tmpDir, "thumbnail.jpg")
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-ss", "0", // Seek to the very beginning
		"-i", videoPath,
		"-frames:v", "1", // Extract exactly 1 frame
		"-vf", "scale=1280:-1", // Scale to 1280px width, maintain aspect ratio
		"-q:v", "2", // High quality
		"-y", thumbnailPath,
	)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to extract thumbnail: %w", err)
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

// generateNarratorVoiceover generates narrator voiceover with variable speed for side effects
func (h *GenerateHandler) generateNarratorVoiceover(
	ctx context.Context,
	userID string,
	jobID string,
	voice string,
	narratorScript string,
	sideEffectsStartTime float64,
	duration int,
) (string, error) {
	if h.ttsAdapter == nil {
		return "", fmt.Errorf("tts adapter not configured")
	}

	if strings.TrimSpace(narratorScript) == "" {
		return "", fmt.Errorf("narrator script is empty")
	}

	h.logger.Info("Generating narrator voiceover",
		zap.String("job_id", jobID),
		zap.String("voice", voice),
		zap.Int("script_length", len(narratorScript)),
		zap.Float64("side_effects_start_time", sideEffectsStartTime),
	)

	audioData, err := h.ttsAdapter.GenerateVoiceover(ctx, narratorScript, voice)
	if err != nil {
		return "", fmt.Errorf("tts generation failed: %w", err)
	}

	h.logger.Info("TTS generation successful",
		zap.String("job_id", jobID),
		zap.Int("audio_size_bytes", len(audioData)),
	)

	tmpDir := filepath.Join("/tmp", jobID, "narrator")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	rawAudioPath := filepath.Join(tmpDir, "narrator-raw.mp3")
	if err := os.WriteFile(rawAudioPath, audioData, 0o644); err != nil {
		return "", fmt.Errorf("failed to write narrator audio: %w", err)
	}

	processedAudioPath := filepath.Join(tmpDir, "narrator-processed.mp3")

	// Get actual audio duration using ffprobe
	audioDuration, err := getAudioDuration(ctx, rawAudioPath)
	if err != nil {
		h.logger.Warn("Failed to get audio duration, skipping variable speed",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		audioDuration = 0
	}

	h.logger.Info("TTS audio generated",
		zap.String("job_id", jobID),
		zap.Float64("audio_duration", audioDuration),
		zap.Float64("video_duration", float64(duration)),
		zap.Float64("requested_split_time", sideEffectsStartTime),
	)

	// Calculate proportional split point based on actual audio duration
	// sideEffectsStartTime is 80% of video duration, we want 80% of audio duration
	var actualSplitTime float64
	if sideEffectsStartTime > 0 && float64(duration) > 0 {
		// Use same ratio for audio: if side effects start at 80% of video, start at 80% of audio
		ratio := sideEffectsStartTime / float64(duration)
		actualSplitTime = audioDuration * ratio
	}

	// Only apply variable speed if we have valid audio and split point
	if actualSplitTime > 0 && audioDuration > 0 && actualSplitTime < audioDuration-0.5 {
		h.logger.Info("Applying variable speed to narrator audio",
			zap.String("job_id", jobID),
			zap.Float64("audio_duration", audioDuration),
			zap.Float64("split_time", actualSplitTime),
		)

		mainSegmentPath := filepath.Join(tmpDir, "main-segment.mp3")
		sideEffectsSegmentPath := filepath.Join(tmpDir, "side-effects.mp3")
		sideEffectsFastPath := filepath.Join(tmpDir, "side-effects-fast.mp3")

		cmd := exec.CommandContext(ctx, "ffmpeg",
			"-i", rawAudioPath,
			"-ss", "0",
			"-to", fmt.Sprintf("%.2f", actualSplitTime),
			"-c", "copy",
			"-y", mainSegmentPath,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to extract main segment: %w (%s)", err, strings.TrimSpace(string(output)))
		}

		cmd = exec.CommandContext(ctx, "ffmpeg",
			"-i", rawAudioPath,
			"-ss", fmt.Sprintf("%.2f", actualSplitTime),
			"-c", "copy",
			"-y", sideEffectsSegmentPath,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to extract side effects segment: %w (%s)", err, strings.TrimSpace(string(output)))
		}

		cmd = exec.CommandContext(ctx, "ffmpeg",
			"-i", sideEffectsSegmentPath,
			"-filter:a", "atempo=1.4",
			"-y", sideEffectsFastPath,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to speed up side effects segment: %w (%s)", err, strings.TrimSpace(string(output)))
		}

		concatFile := filepath.Join(tmpDir, "concat.txt")
		concatContent := fmt.Sprintf("file '%s'\nfile '%s'\n", mainSegmentPath, sideEffectsFastPath)
		if err := os.WriteFile(concatFile, []byte(concatContent), 0o644); err != nil {
			return "", fmt.Errorf("failed to create concat file: %w", err)
		}

		cmd = exec.CommandContext(ctx, "ffmpeg",
			"-f", "concat",
			"-safe", "0",
			"-i", concatFile,
			"-c", "copy",
			"-y", processedAudioPath,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to concatenate narrator segments: %w (%s)", err, strings.TrimSpace(string(output)))
		}
	} else {
		h.logger.Info("Skipping variable speed for narrator audio",
			zap.String("job_id", jobID),
			zap.Float64("side_effects_start_time", sideEffectsStartTime),
			zap.Float64("audio_duration", audioDuration),
			zap.Float64("actual_split_time", actualSplitTime),
		)

		if err := os.Rename(rawAudioPath, processedAudioPath); err != nil {
			return "", fmt.Errorf("failed to finalize narrator audio: %w", err)
		}
	}

	h.logger.Info("Uploading narrator audio to S3", zap.String("job_id", jobID))
	s3Key := buildNarratorAudioKey(userID, jobID)
	narratorAudioURL, err := h.s3Service.UploadFile(ctx, h.assetsBucket, s3Key, processedAudioPath, "audio/mpeg")
	if err != nil {
		return "", fmt.Errorf("failed to upload narrator audio: %w", err)
	}

	h.logger.Info("Narrator voiceover uploaded",
		zap.String("job_id", jobID),
		zap.String("s3_key", s3Key),
		zap.String("url", narratorAudioURL),
	)

	return narratorAudioURL, nil
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

// getAudioDuration returns the duration of an audio file in seconds using ffprobe
func getAudioDuration(ctx context.Context, audioPath string) (float64, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath,
	)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration '%s': %w", durationStr, err)
	}

	return duration, nil
}

// detectAvailableFont returns the first available font file path from a prioritized list.
func detectAvailableFont(logger *zap.Logger) string {
	fontPaths := []string{
		// Alpine Linux (Docker container)
		"/usr/share/fonts/ttf-dejavu/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/ttf-dejavu/DejaVuSans.ttf",
		// Debian/Ubuntu
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf",
		"/usr/share/fonts/truetype/liberation2/LiberationSans-Bold.ttf",
	}

	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			logger.Info("Font detected for text overlay", zap.String("font_path", path))
			return path
		}
	}

	logger.Warn("No preferred fonts found, using ffmpeg default font")
	return ""
}

// escapeFfmpegText escapes special characters so they are safe for ffmpeg drawtext filters.
func escapeFfmpegText(text string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\", // Backslashes (must be first)
		"'", "\\'", // Single quotes
		":", "\\:", // Colons
		"%", "\\%", // Percent signs
		"\n", "\\n", // Newlines → \n
		"\"", "\\\"", // Double quotes
	)
	return replacer.Replace(text)
}

type drawtextConfig struct {
	Filter         string
	OverlayStart   float64
	OverlayEnd     float64
	RuneCount      int
	BaseFontSize   float64
	MaxChars       int
	EstimatedWidth float64
	RenderedText   string
}

func buildDrawtextConfig(
	logger *zap.Logger,
	text string,
	sideEffectsStartTime float64,
	totalDuration float64,
	videoWidth int,
	videoHeight int,
) (*drawtextConfig, error) {
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" {
		return nil, nil
	}

	if totalDuration <= 0 {
		logger.Warn("Skipping text overlay (unknown video duration)",
			zap.Float64("video_duration", totalDuration),
		)
		return nil, nil
	}

	overlayStart := sideEffectsStartTime
	if overlayStart <= 0 {
		overlayStart = totalDuration * 0.8
		logger.Warn("Non-positive side effects start time; defaulting to last 20%",
			zap.Float64("original_start_time", sideEffectsStartTime),
			zap.Float64("fallback_start_time", overlayStart),
		)
	}

	overlayEnd := totalDuration
	if overlayStart >= overlayEnd {
		logger.Warn("Skipping text overlay (start time beyond duration)",
			zap.Float64("side_effects_start_time", overlayStart),
			zap.Float64("video_duration", overlayEnd),
		)
		return nil, nil
	}

	runeCount := utf8.RuneCountInString(trimmedText)
	baseFontSize := 36.0
	if runeCount > 360 {
		baseFontSize = 32.0
	}
	if runeCount > 440 {
		baseFontSize = 28.0
	}

	if videoWidth <= 0 {
		videoWidth = 1920
	}
	if videoHeight <= 0 {
		videoHeight = 1080
	}

	scaleFactor := float64(videoHeight) / 1080.0
	fontSizePixels := baseFontSize * scaleFactor
	if fontSizePixels < 18 {
		fontSizePixels = 18
	}

	lineSpacingPixels := int(math.Round(fontSizePixels * (8.0 / 36.0)))
	wrappedText, maxChars := wrapText(trimmedText, videoWidth, fontSizePixels)

	estimatedWidthPx := float64(maxChars) * fontSizePixels * 0.6
	if estimatedWidthPx > float64(videoWidth) {
		estimatedWidthPx = float64(videoWidth)
	}

	fontFile := detectAvailableFont(logger)
	escapedText := escapeFfmpegText(wrappedText)

	filterParts := []string{
		fmt.Sprintf("text='%s'", escapedText),
		fmt.Sprintf("fontsize=%.2f", fontSizePixels),
	}
	if fontFile != "" {
		filterParts = append(filterParts, fmt.Sprintf("fontfile='%s'", fontFile))
	}
	filterParts = append(filterParts,
		"fontcolor=white",
		"bordercolor=black",
		"borderw=2",
		fmt.Sprintf("x=(w-%.2f)/2", estimatedWidthPx),
		"y=h-h*0.2",
		fmt.Sprintf("line_spacing=%d", lineSpacingPixels),
		fmt.Sprintf("enable='between(t,%.2f,%.2f)'", overlayStart, overlayEnd),
	)

	return &drawtextConfig{
		Filter:         fmt.Sprintf("drawtext=%s", strings.Join(filterParts, ":")),
		OverlayStart:   overlayStart,
		OverlayEnd:     overlayEnd,
		RuneCount:      runeCount,
		BaseFontSize:   baseFontSize,
		MaxChars:       maxChars,
		EstimatedWidth: estimatedWidthPx,
		RenderedText:   wrappedText,
	}, nil
}

func wrapText(text string, videoWidth int, fontSize float64) (string, int) {
	totalRunes := utf8.RuneCountInString(text)
	if totalRunes <= 60 {
		estimatedWidth := float64(totalRunes) * fontSize * 0.6
		if estimatedWidth <= float64(videoWidth)*0.8 {
			return text, totalRunes
		}
	}

	maxCharsPerLine := int((float64(videoWidth) * 0.8) / (fontSize * 0.6))
	if maxCharsPerLine >= totalRunes {
		return text, totalRunes
	}
	if maxCharsPerLine < 20 {
		maxCharsPerLine = 20
	}

	lines := make([]string, 0)
	for _, rawLine := range strings.Split(text, "\n") {
		words := strings.Fields(rawLine)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		var currentWords []string
		currentLen := 0
		for _, word := range words {
			wordLen := utf8.RuneCountInString(word)
			additionalLen := wordLen
			if currentLen > 0 {
				additionalLen++ // account for space
			}

			if currentLen+additionalLen > maxCharsPerLine {
				lines = append(lines, strings.Join(currentWords, " "))
				currentWords = []string{word}
				currentLen = wordLen
			} else {
				if currentLen > 0 {
					currentLen++ // space
				}
				currentWords = append(currentWords, word)
				currentLen += wordLen
			}
		}

		if len(currentWords) > 0 {
			lines = append(lines, strings.Join(currentWords, " "))
		}
	}

	return strings.Join(lines, "\n"), maxCharsPerLine
}

// composeVideo concatenates video clips and prepares the final video track (no audio mixing).
// 4-Track Architecture:
//   - Video Track: final/video.mp4 (output of this function)
//   - Music Track: audio/background-music.mp3 (generated separately)
//   - Narrator Track: audio/narrator-voiceover.mp3 (generated separately)
//   - Text Track: side_effects_text metadata (rendered by the frontend)
//
// All audio playback, mixing, and synchronization happens in the frontend.
func (h *GenerateHandler) composeVideo(
	ctx context.Context,
	userID string,
	jobID string,
	clips []ClipVideo,
	sideEffectsText string,
	sideEffectsStartTime float64,
) (string, error) {
	trimmedText := strings.TrimSpace(sideEffectsText)
	var totalDuration float64
	for _, clip := range clips {
		totalDuration += clip.Duration
	}

	h.logger.Info("Composing final video",
		zap.String("job_id", jobID),
		zap.Int("num_clips", len(clips)),
		zap.Bool("has_side_effects", trimmedText != ""),
		zap.Float64("side_effects_start_time", sideEffectsStartTime),
		zap.Float64("video_duration_estimate", totalDuration),
	)

	if sideEffectsStartTime > 0 && trimmedText == "" {
		return "", fmt.Errorf("side effects text is required when sideEffectsStartTime is provided")
	}

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

	// Create concat file for ffmpeg
	concatFile := filepath.Join(tmpDir, "concat.txt")
	f, err := os.Create(concatFile)
	if err != nil {
		return "", fmt.Errorf("failed to create concat file: %w", err)
	}
	for _, path := range clipPaths {
		fmt.Fprintf(f, "file '%s'\n", path)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("failed to close concat file: %w", err)
	}

	// Concatenate clips (video track only)
	h.logger.Info("Concatenating video clips (video track only)",
		zap.String("job_id", jobID),
		zap.Int("num_clips", len(clipPaths)),
	)
	finalVideo := filepath.Join(tmpDir, "final.mp4")
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c:v", "copy",
		"-an", // Explicitly drop audio streams (frontend handles audio tracks)
		"-y", finalVideo,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		h.logger.Error("ffmpeg concat failed",
			zap.String("job_id", jobID),
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return "", fmt.Errorf("ffmpeg concat failed: %w", err)
	}

	if trimmedText != "" && totalDuration > 0 {
		videoWidth, videoHeight, err := probeVideoDimensions(finalVideo)
		if err != nil {
			h.logger.Warn("Failed to probe video dimensions, using defaults",
				zap.String("job_id", jobID),
				zap.Error(err),
			)
			videoWidth = 1920
			videoHeight = 1080
		}

		config, err := buildDrawtextConfig(h.logger, trimmedText, sideEffectsStartTime, totalDuration, videoWidth, videoHeight)
		if err != nil {
			return "", err
		}
		if config != nil {
			h.logger.Info("Applying side effects text overlay",
				zap.String("job_id", jobID),
				zap.Float64("overlay_start", config.OverlayStart),
				zap.Float64("overlay_end", config.OverlayEnd),
				zap.Int("rune_count", config.RuneCount),
				zap.Float64("base_font_size", config.BaseFontSize),
			)

			videoWithText := filepath.Join(tmpDir, "video_with_text.mp4")
			cmd = exec.CommandContext(ctx, "ffmpeg",
				"-i", finalVideo,
				"-vf", config.Filter,
				"-c:v", "libx264",
				"-preset", "medium",
				"-crf", "21",
				"-y", videoWithText,
			)
			if output, err := cmd.CombinedOutput(); err != nil {
				h.logger.Error("ffmpeg text overlay failed",
					zap.String("job_id", jobID),
					zap.String("output", string(output)),
					zap.Error(err),
				)
				return "", fmt.Errorf("ffmpeg text overlay failed: %w", err)
			}

			h.logger.Info("Text overlay applied successfully", zap.String("job_id", jobID))
			finalVideo = videoWithText
		}
	} else if trimmedText == "" {
		h.logger.Info("Skipping text overlay (no side effects text)", zap.String("job_id", jobID))
	} else {
		h.logger.Warn("Skipping text overlay (unknown video duration)",
			zap.String("job_id", jobID),
			zap.Float64("video_duration", totalDuration),
		)
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
	// Strip query parameters first (handles presigned URLs stored in DB)
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

func probeVideoDimensions(videoPath string) (int, int, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0:s=x",
		videoPath,
	)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	var width, height int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%dx%d", &width, &height); err != nil {
		return 0, 0, fmt.Errorf("failed to parse ffprobe dimensions: %w", err)
	}
	return width, height, nil
}
