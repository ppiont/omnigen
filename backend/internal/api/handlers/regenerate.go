package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/adapters"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// RegenerateHandler handles scene regeneration requests
type RegenerateHandler struct {
	jobRepo      *repository.DynamoDBRepository
	s3Service    *repository.S3AssetRepository
	veoAdapter   *adapters.VeoAdapter
	assetsBucket string
	logger       *zap.Logger
}

// NewRegenerateHandler creates a new regenerate handler
func NewRegenerateHandler(
	jobRepo *repository.DynamoDBRepository,
	s3Service *repository.S3AssetRepository,
	veoAdapter *adapters.VeoAdapter,
	assetsBucket string,
	logger *zap.Logger,
) *RegenerateHandler {
	return &RegenerateHandler{
		jobRepo:      jobRepo,
		s3Service:    s3Service,
		veoAdapter:   veoAdapter,
		assetsBucket: assetsBucket,
		logger:       logger,
	}
}

// RegenerateRequest represents a scene regeneration request
type RegenerateRequest struct {
	Cascade bool `json:"cascade" form:"cascade"` // Regenerate subsequent scenes too
}

// RegenerateResponse represents a scene regeneration response
type RegenerateResponse struct {
	JobID        string `json:"job_id"`
	SceneNumber  int    `json:"scene_number"`
	NewVersion   int    `json:"new_version"`
	ClipURL      string `json:"clip_url"`
	CascadeCount int    `json:"cascade_count,omitempty"` // How many subsequent scenes regenerated
}

// RegenerateScene handles POST /api/v1/jobs/:id/scenes/:scene_number/regenerate
// @Summary Regenerate a specific scene
// @Description Regenerates a single scene of a completed job with versioning support
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param scene_number path int true "Scene number (1-indexed)"
// @Param request body RegenerateRequest false "Regeneration options"
// @Success 200 {object} RegenerateResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/jobs/{id}/scenes/{scene_number}/regenerate [post]
// @Security BearerAuth
func (h *RegenerateHandler) RegenerateScene(c *gin.Context) {
	jobID := c.Param("id")
	sceneNumStr := c.Param("scene_number")
	userID := auth.MustGetUserID(c)

	sceneNum, err := strconv.Atoi(sceneNumStr)
	if err != nil || sceneNum < 1 {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.NewValidationError("scene_number", "Invalid scene number"),
		})
		return
	}

	var req RegenerateRequest
	// Bind query params and body, ignore errors for optional params
	_ = c.ShouldBindQuery(&req)
	_ = c.ShouldBindJSON(&req)

	h.logger.Info("Scene regeneration requested",
		zap.String("job_id", jobID),
		zap.Int("scene_number", sceneNum),
		zap.Bool("cascade", req.Cascade),
		zap.String("user_id", userID),
	)

	// Fetch job and validate
	job, err := h.jobRepo.GetJob(c.Request.Context(), jobID)
	if err != nil {
		if err == repository.ErrJobNotFound {
			c.JSON(http.StatusNotFound, errors.ErrorResponse{
				Error: errors.ErrJobNotFound,
			})
			return
		}
		h.logger.Error("Failed to get job", zap.String("job_id", jobID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrDatabaseError,
		})
		return
	}

	// Verify job belongs to the current user
	if job.UserID != userID {
		c.JSON(http.StatusNotFound, errors.ErrorResponse{
			Error: errors.ErrJobNotFound,
		})
		return
	}

	// Can only regenerate scenes from completed jobs
	if job.Status != domain.StatusCompleted {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.NewValidationError("status", "Can only regenerate scenes from completed jobs"),
		})
		return
	}

	// Validate scene number
	if sceneNum > len(job.Scenes) || sceneNum < 1 {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.NewValidationError("scene_number", fmt.Sprintf("Invalid scene number. Job has %d scenes.", len(job.Scenes))),
		})
		return
	}

	// Get start image from previous scene's last frame (or nothing for scene 1)
	var startImageURL string
	if sceneNum > 1 {
		// Get the previous scene's thumbnail as start image
		prevSceneNum := sceneNum - 1
		prevVersion := 1
		if job.SceneVersions != nil && job.SceneVersions[prevSceneNum] > 0 {
			prevVersion = job.SceneVersions[prevSceneNum]
		}

		// Try versioned thumbnail first, then fall back to non-versioned
		prevThumbnailKey := buildVersionedSceneThumbnailKey(job.UserID, jobID, prevSceneNum, prevVersion)
		presignedURL, err := h.s3Service.GetPresignedURL(c.Request.Context(), prevThumbnailKey, 1*time.Hour)
		if err != nil {
			// Fall back to non-versioned thumbnail
			prevThumbnailKey = buildSceneThumbnailKey(job.UserID, jobID, prevSceneNum)
			presignedURL, err = h.s3Service.GetPresignedURL(c.Request.Context(), prevThumbnailKey, 1*time.Hour)
			if err != nil {
				h.logger.Warn("Could not get previous scene thumbnail for continuity",
					zap.String("job_id", jobID),
					zap.Int("prev_scene", prevSceneNum),
					zap.Error(err),
				)
			}
		}
		if err == nil {
			startImageURL = presignedURL
		}
	}

	// Get scene metadata from stored script
	scene := job.Scenes[sceneNum-1]
	scene.StartImageURL = startImageURL

	// Generate new clip
	ctx := c.Request.Context()
	clipResult, err := h.generateClip(ctx, job.UserID, jobID, scene, job.AspectRatio, sceneNum)
	if err != nil {
		h.logger.Error("Scene regeneration failed",
			zap.String("job_id", jobID),
			zap.Int("scene_number", sceneNum),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer.WithDetails(map[string]interface{}{
				"message": "Scene regeneration failed",
				"error":   err.Error(),
			}),
		})
		return
	}

	// Initialize version maps if needed
	if job.SceneVersions == nil {
		job.SceneVersions = make(map[int]int)
	}
	if job.ClipVersions == nil {
		job.ClipVersions = make(map[string]string)
	}

	// Increment version
	currentVersion := job.SceneVersions[sceneNum]
	newVersion := currentVersion + 1

	// Store versioned clip
	versionKey := fmt.Sprintf("scene-%d-v%d", sceneNum, newVersion)
	job.SceneVersions[sceneNum] = newVersion
	job.ClipVersions[versionKey] = clipResult.VideoURL

	// Update current scene URL in array
	job.SceneVideoURLs[sceneNum-1] = clipResult.VideoURL

	// Handle cascade regeneration if requested
	cascadeCount := 0
	if req.Cascade && sceneNum < len(job.Scenes) {
		h.logger.Info("Cascade regeneration requested",
			zap.String("job_id", jobID),
			zap.Int("starting_scene", sceneNum+1),
			zap.Int("total_scenes", len(job.Scenes)),
		)

		// Use the new scene's last frame as start image for next scene
		nextStartImageURL := clipResult.LastFrameURL

		for nextScene := sceneNum + 1; nextScene <= len(job.Scenes); nextScene++ {
			nextSceneData := job.Scenes[nextScene-1]
			nextSceneData.StartImageURL = nextStartImageURL

			nextClipResult, err := h.generateClip(ctx, job.UserID, jobID, nextSceneData, job.AspectRatio, nextScene)
			if err != nil {
				h.logger.Error("Cascade scene regeneration failed",
					zap.String("job_id", jobID),
					zap.Int("scene_number", nextScene),
					zap.Error(err),
				)
				// Continue with partial success
				break
			}

			// Update version for cascaded scene
			nextCurrentVersion := job.SceneVersions[nextScene]
			nextNewVersion := nextCurrentVersion + 1
			nextVersionKey := fmt.Sprintf("scene-%d-v%d", nextScene, nextNewVersion)
			job.SceneVersions[nextScene] = nextNewVersion
			job.ClipVersions[nextVersionKey] = nextClipResult.VideoURL
			job.SceneVideoURLs[nextScene-1] = nextClipResult.VideoURL

			nextStartImageURL = nextClipResult.LastFrameURL
			cascadeCount++
		}
	}

	// Re-compose final video with updated clips
	clips := h.buildClipVideosFromJob(job)
	mp4Key, webmKey, err := h.composeVideo(ctx, job, clips)
	if err != nil {
		h.logger.Error("Video recomposition failed",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer.WithDetails(map[string]interface{}{
				"message": "Video recomposition failed",
				"error":   err.Error(),
			}),
		})
		return
	}

	job.VideoKey = mp4Key
	job.WebMVideoKey = webmKey
	job.UpdatedAt = time.Now().Unix()

	// Save updated job
	if err := h.jobRepo.UpdateJob(ctx, job); err != nil {
		h.logger.Error("Failed to update job after regeneration",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrDatabaseError,
		})
		return
	}

	// Generate presigned URL for the new clip
	clipPresignedURL, err := h.s3Service.GetPresignedURL(ctx, extractS3Key(clipResult.VideoURL), 7*24*time.Hour)
	if err != nil {
		clipPresignedURL = clipResult.VideoURL
	}

	h.logger.Info("Scene regeneration complete",
		zap.String("job_id", jobID),
		zap.Int("scene_number", sceneNum),
		zap.Int("new_version", newVersion),
		zap.Int("cascade_count", cascadeCount),
	)

	c.JSON(http.StatusOK, RegenerateResponse{
		JobID:        jobID,
		SceneNumber:  sceneNum,
		NewVersion:   newVersion,
		ClipURL:      clipPresignedURL,
		CascadeCount: cascadeCount,
	})
}

// generateClip generates a single video clip using Veo adapter
// This is a simplified version that reuses logic from generate_async.go
func (h *RegenerateHandler) generateClip(
	ctx context.Context,
	userID string,
	jobID string,
	scene domain.Scene,
	aspectRatio string,
	clipNumber int,
) (ClipVideo, error) {
	h.logger.Info("Regenerating scene clip",
		zap.String("job_id", jobID),
		zap.Int("scene", clipNumber),
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
	maxAttempts := VideoGenerationMaxAttempts
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
			// Process and upload the video
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
			errorMsg := result.Error
			if errorMsg == "" {
				errorMsg = "Unknown error - Veo returned failed status"
			}
			return ClipVideo{}, fmt.Errorf("veo generation failed: %s", errorMsg)
		}

		// Log progress periodically
		if attempt%12 == 0 {
			h.logger.Info("Veo still processing",
				zap.String("job_id", jobID),
				zap.Int("attempt", attempt),
				zap.String("status", result.Status),
			)
		}
	}

	return ClipVideo{}, fmt.Errorf("clip generation timed out after %d attempts", maxAttempts)
}

// processVideo downloads video from Replicate, extracts last frame, uploads both to S3
func (h *RegenerateHandler) processVideo(
	ctx context.Context,
	userID string,
	jobID string,
	clipNumber int,
	videoURL string,
) (string, string, error) {
	return processVideoCommon(ctx, h.s3Service, h.assetsBucket, h.logger, userID, jobID, clipNumber, videoURL)
}

// buildClipVideosFromJob constructs ClipVideo slice from job data
func (h *RegenerateHandler) buildClipVideosFromJob(job *domain.Job) []ClipVideo {
	clips := make([]ClipVideo, len(job.SceneVideoURLs))
	for i, url := range job.SceneVideoURLs {
		duration := 8.0 // Default duration
		if i < len(job.Scenes) {
			duration = job.Scenes[i].Duration
		}
		clips[i] = ClipVideo{
			VideoURL: url,
			Duration: duration,
		}
	}
	return clips
}

// composeVideo recomposes the final video from all clips
// Returns: (mp4Key, webmKey, error) - webmKey may be empty if WebM encoding fails
func (h *RegenerateHandler) composeVideo(
	ctx context.Context,
	job *domain.Job,
	clips []ClipVideo,
) (string, string, error) {
	return composeVideoCommon(ctx, h.s3Service, h.assetsBucket, h.logger, job, clips)
}
