package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/internal/service"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// JobsHandler handles job-related requests
type JobsHandler struct {
	jobRepo      *repository.DynamoDBRepository
	s3Service    *repository.S3AssetRepository
	assetService *service.AssetService
	assetsBucket string
	logger       *zap.Logger
}

// NewJobsHandler creates a new jobs handler
func NewJobsHandler(
	jobRepo *repository.DynamoDBRepository,
	s3Service *repository.S3AssetRepository,
	assetService *service.AssetService,
	assetsBucket string,
	logger *zap.Logger,
) *JobsHandler {
	return &JobsHandler{
		jobRepo:      jobRepo,
		s3Service:    s3Service,
		assetService: assetService,
		assetsBucket: assetsBucket,
		logger:       logger,
	}
}

// JobResponse represents a job status response
type JobResponse struct {
	JobID           string  `json:"job_id"`
	Status          string  `json:"status"`
	Stage           string  `json:"stage,omitempty"`
	ProgressPercent int     `json:"progress_percent"`
	Prompt          string  `json:"prompt"`
	Duration        int     `json:"duration"`
	VideoURL        *string `json:"video_url,omitempty"`
	CreatedAt       int64   `json:"created_at"`
	UpdatedAt       int64   `json:"updated_at"`
	CompletedAt     *int64  `json:"completed_at,omitempty"`
	ErrorMessage    *string `json:"error_message,omitempty"`

	// Progress fields
	ThumbnailURL    string   `json:"thumbnail_url,omitempty"`
	AudioURL        string   `json:"audio_url,omitempty"`
	ScenesCompleted int      `json:"scenes_completed,omitempty"`
	SceneVideoURLs  []string `json:"scene_video_urls,omitempty"`
}

// ListJobsResponse represents a list of jobs
type ListJobsResponse struct {
	Jobs       []JobResponse `json:"jobs"`
	TotalCount int           `json:"total_count"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
}

// GetJob handles GET /api/v1/jobs/:id
// @Summary Get job status
// @Description Get the status and details of a video generation job
// @Tags jobs
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} JobResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/jobs/{id} [get]
// @Security BearerAuth
func (h *JobsHandler) GetJob(c *gin.Context) {
	jobID := c.Param("id")

	// Get job from DynamoDB
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

	// Generate presigned URL if video is completed
	var videoURL *string
	if job.Status == "completed" && job.VideoKey != "" {
		url, err := h.s3Service.GetPresignedURL(c.Request.Context(), job.VideoKey, 7*24*time.Hour)
		if err != nil {
			h.logger.Error("Failed to generate presigned URL",
				zap.String("job_id", jobID),
				zap.Error(err),
			)
		} else {
			videoURL = &url
		}
	}

	// Generate presigned URL for thumbnail
	var thumbnailURL string
	if job.ThumbnailURL != "" {
		key := extractS3Key(job.ThumbnailURL)
		url, err := h.s3Service.GetPresignedURL(c.Request.Context(), key, 7*24*time.Hour)
		if err != nil {
			h.logger.Warn("Failed to generate presigned URL for thumbnail",
				zap.String("job_id", jobID),
				zap.Error(err),
			)
			thumbnailURL = job.ThumbnailURL // Fallback to raw URL
		} else {
			thumbnailURL = url
		}
	}

	// Generate presigned URL for audio
	var audioURL string
	if job.AudioURL != "" {
		key := extractS3Key(job.AudioURL)
		url, err := h.s3Service.GetPresignedURL(c.Request.Context(), key, 7*24*time.Hour)
		if err != nil {
			h.logger.Warn("Failed to generate presigned URL for audio",
				zap.String("job_id", jobID),
				zap.Error(err),
			)
			audioURL = job.AudioURL // Fallback to raw URL
		} else {
			audioURL = url
		}
	}

	// Generate presigned URLs for scene videos
	var sceneVideoURLs []string
	if len(job.SceneVideoURLs) > 0 {
		sceneVideoURLs = make([]string, 0, len(job.SceneVideoURLs))
		for i, sceneURL := range job.SceneVideoURLs {
			if sceneURL == "" {
				continue
			}
			key := extractS3Key(sceneURL)
			url, err := h.s3Service.GetPresignedURL(c.Request.Context(), key, 7*24*time.Hour)
			if err != nil {
				h.logger.Warn("Failed to generate presigned URL for scene video",
					zap.String("job_id", jobID),
					zap.Int("scene", i+1),
					zap.Error(err),
				)
				sceneVideoURLs = append(sceneVideoURLs, sceneURL) // Fallback to raw URL
			} else {
				sceneVideoURLs = append(sceneVideoURLs, url)
			}
		}
	}

	response := JobResponse{
		JobID:           job.JobID,
		Status:          job.Status,
		Stage:           job.Stage,
		ProgressPercent: calculateDynamicProgress(job.Stage, len(job.Scenes)),
		Prompt:          job.Prompt,
		Duration:        job.Duration,
		VideoURL:        videoURL,
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       job.UpdatedAt,
		CompletedAt:     job.CompletedAt,
		ErrorMessage:    job.ErrorMessage,
		ThumbnailURL:    thumbnailURL,
		AudioURL:        audioURL,
		ScenesCompleted: job.ScenesCompleted,
		SceneVideoURLs:  sceneVideoURLs,
	}

	c.JSON(http.StatusOK, response)
}

// ListJobs handles GET /api/v1/jobs
// @Summary List jobs
// @Description Get a list of video generation jobs
// @Tags jobs
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Success 200 {object} ListJobsResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/jobs [get]
// @Security BearerAuth
func (h *JobsHandler) ListJobs(c *gin.Context) {
	// Get user ID from auth context
	userID := auth.MustGetUserID(c)

	// Get query parameters
	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get optional status filter
	status := c.Query("status")

	h.logger.Info("Listing jobs",
		zap.String("user_id", userID),
		zap.Int("page_size", pageSize),
		zap.String("status_filter", status),
	)

	// Get jobs for user with optional status filter
	jobs, err := h.jobRepo.GetJobsByUser(c.Request.Context(), userID, pageSize, status)
	if err != nil {
		h.logger.Error("Failed to list jobs",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	// Convert to response format
	jobResponses := make([]JobResponse, len(jobs))
	for i, job := range jobs {
		// Convert VideoKey to presigned URL if present
		var videoURL *string
		if job.VideoKey != "" {
			url, err := h.s3Service.GetPresignedURL(c.Request.Context(), job.VideoKey, 1*time.Hour)
			if err != nil {
				h.logger.Warn("Failed to generate presigned URL",
					zap.String("job_id", job.JobID),
					zap.String("video_key", job.VideoKey),
					zap.Error(err),
				)
			} else {
				videoURL = &url
			}
		}

		// Generate presigned URL for thumbnail
		var thumbnailURL string
		if job.ThumbnailURL != "" {
			key := extractS3Key(job.ThumbnailURL)
			url, err := h.s3Service.GetPresignedURL(c.Request.Context(), key, 1*time.Hour)
			if err != nil {
				h.logger.Warn("Failed to generate presigned URL for thumbnail",
					zap.String("job_id", job.JobID),
					zap.Error(err),
				)
				thumbnailURL = job.ThumbnailURL // Fallback to raw URL
			} else {
				thumbnailURL = url
			}
		}

		// Generate presigned URL for audio
		var audioURL string
		if job.AudioURL != "" {
			key := extractS3Key(job.AudioURL)
			url, err := h.s3Service.GetPresignedURL(c.Request.Context(), key, 1*time.Hour)
			if err != nil {
				h.logger.Warn("Failed to generate presigned URL for audio",
					zap.String("job_id", job.JobID),
					zap.Error(err),
				)
				audioURL = job.AudioURL // Fallback to raw URL
			} else {
				audioURL = url
			}
		}

		// Generate presigned URLs for scene videos
		var sceneVideoURLs []string
		if len(job.SceneVideoURLs) > 0 {
			sceneVideoURLs = make([]string, 0, len(job.SceneVideoURLs))
			for si, sceneURL := range job.SceneVideoURLs {
				if sceneURL == "" {
					continue
				}
				key := extractS3Key(sceneURL)
				url, err := h.s3Service.GetPresignedURL(c.Request.Context(), key, 1*time.Hour)
				if err != nil {
					h.logger.Warn("Failed to generate presigned URL for scene video",
						zap.String("job_id", job.JobID),
						zap.Int("scene", si+1),
						zap.Error(err),
					)
					sceneVideoURLs = append(sceneVideoURLs, sceneURL) // Fallback to raw URL
				} else {
					sceneVideoURLs = append(sceneVideoURLs, url)
				}
			}
		}

		jobResponses[i] = JobResponse{
			JobID:           job.JobID,
			Status:          job.Status,
			Stage:           job.Stage,
			ProgressPercent: calculateDynamicProgress(job.Stage, len(job.Scenes)),
			VideoURL:        videoURL,
			ErrorMessage:    job.ErrorMessage,
			Prompt:          job.Prompt,
			Duration:        job.Duration,
			CreatedAt:       job.CreatedAt,
			UpdatedAt:       job.UpdatedAt,
			CompletedAt:     job.CompletedAt,
			ThumbnailURL:    thumbnailURL,
			AudioURL:        audioURL,
			ScenesCompleted: job.ScenesCompleted,
			SceneVideoURLs:  sceneVideoURLs,
		}
	}

	response := ListJobsResponse{
		Jobs:       jobResponses,
		TotalCount: len(jobResponses),
		Page:       1,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteJob handles DELETE /api/v1/jobs/:id
// @Summary Delete a job
// @Description Delete a video generation job and its associated assets from DynamoDB and S3
// @Tags jobs
// @Param id path string true "Job ID"
// @Success 204 "No Content"
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/jobs/{id} [delete]
// @Security BearerAuth
func (h *JobsHandler) DeleteJob(c *gin.Context) {
	jobID := c.Param("id")
	userID := auth.MustGetUserID(c)

	// Get job to verify it exists and get S3 keys
	job, err := h.jobRepo.GetJob(c.Request.Context(), jobID)
	if err != nil {
		if err == repository.ErrJobNotFound {
			c.JSON(http.StatusNotFound, errors.ErrorResponse{
				Error: errors.ErrJobNotFound,
			})
			return
		}

		h.logger.Error("Failed to get job for deletion",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrDatabaseError,
		})
		return
	}

	// Delete S3 assets
	h.logger.Info("Deleting S3 assets for job",
		zap.String("job_id", jobID),
		zap.String("user_id", userID),
	)

	// Delete final video
	if job.VideoKey != "" {
		if err := h.s3Service.DeleteFile(c.Request.Context(), h.assetsBucket, job.VideoKey); err != nil {
			h.logger.Warn("Failed to delete final video from S3",
				zap.String("job_id", jobID),
				zap.String("video_key", job.VideoKey),
				zap.Error(err),
			)
			// Continue with deletion even if S3 delete fails
		}
	}

	// Delete scene videos
	for i, sceneURL := range job.SceneVideoURLs {
		if sceneURL != "" {
			key := extractS3Key(sceneURL)
			if key != "" {
				if err := h.s3Service.DeleteFile(c.Request.Context(), h.assetsBucket, key); err != nil {
					h.logger.Warn("Failed to delete scene video from S3",
						zap.String("job_id", jobID),
						zap.Int("scene", i+1),
						zap.String("key", key),
						zap.Error(err),
					)
					// Continue with deletion even if S3 delete fails
				}
			}
		}
	}

	// Delete thumbnail
	if job.ThumbnailURL != "" {
		key := extractS3Key(job.ThumbnailURL)
		if key != "" {
			if err := h.s3Service.DeleteFile(c.Request.Context(), h.assetsBucket, key); err != nil {
				h.logger.Warn("Failed to delete thumbnail from S3",
					zap.String("job_id", jobID),
					zap.String("key", key),
					zap.Error(err),
				)
				// Continue with deletion even if S3 delete fails
			}
		}
	}

	// Delete audio
	if job.AudioURL != "" {
		key := extractS3Key(job.AudioURL)
		if key != "" {
			if err := h.s3Service.DeleteFile(c.Request.Context(), h.assetsBucket, key); err != nil {
				h.logger.Warn("Failed to delete audio from S3",
					zap.String("job_id", jobID),
					zap.String("key", key),
					zap.Error(err),
				)
				// Continue with deletion even if S3 delete fails
			}
		}
	}

	// Delete job from DynamoDB
	err = h.jobRepo.DeleteJob(c.Request.Context(), jobID)
	if err != nil {
		h.logger.Error("Failed to delete job from DynamoDB",
			zap.String("job_id", jobID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	h.logger.Info("Job and assets deleted successfully",
		zap.String("job_id", jobID),
		zap.String("user_id", userID),
	)

	c.Status(http.StatusNoContent)
}

// calculateDynamicProgress calculates progress percentage based on stage and total scenes
// Progress allocation: Script (15%), Scenes (70%), Audio (10%), Composition (5%)
// NOTE: This is duplicated from jobs_stream.go - could be refactored to shared package
func calculateDynamicProgress(stage string, totalScenes int) int {
	// Handle edge case
	if totalScenes == 0 {
		totalScenes = 3 // Default to 3 scenes if not set
	}

	// Script generation stages
	if stage == "script_generating" {
		return 5
	}
	if stage == "script_complete" {
		return 15
	}

	// Scene generation stages (70% allocated, split evenly)
	if len(stage) > 6 && stage[:6] == "scene_" {
		var sceneNum int
		var suffix string
		_, err := fmt.Sscanf(stage, "scene_%d_%s", &sceneNum, &suffix)
		if err == nil && sceneNum > 0 && sceneNum <= totalScenes {
			percentPerScene := 70.0 / float64(totalScenes)

			switch suffix {
			case "generating":
				progress := 15.0 + float64(sceneNum-1)*percentPerScene
				return int(progress)
			case "complete":
				progress := 15.0 + float64(sceneNum)*percentPerScene
				return int(progress)
			}
		}
	}

	// Audio generation stages
	if stage == "audio_generating" {
		return 85
	}
	if stage == "audio_complete" {
		return 95
	}

	// Composition stage
	if stage == "composing" {
		return 95
	}

	// Completion
	if stage == "complete" {
		return 100
	}

	return 0
}
