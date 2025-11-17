package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// JobsHandler handles job-related requests
type JobsHandler struct {
	jobRepo   *repository.DynamoDBRepository
	s3Service *repository.S3AssetRepository
	logger    *zap.Logger
}

// NewJobsHandler creates a new jobs handler
func NewJobsHandler(
	jobRepo *repository.DynamoDBRepository,
	s3Service *repository.S3AssetRepository,
	logger *zap.Logger,
) *JobsHandler {
	return &JobsHandler{
		jobRepo:   jobRepo,
		s3Service: s3Service,
		logger:    logger,
	}
}

// JobResponse represents a job status response
type JobResponse struct {
	JobID           string                 `json:"job_id"`
	Status          string                 `json:"status"`
	Stage           string                 `json:"stage,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ProgressPercent int                    `json:"progress_percent"`
	Prompt          string                 `json:"prompt"`
	Duration        int                    `json:"duration"`
	VideoURL        *string                `json:"video_url,omitempty"`
	CreatedAt       int64                  `json:"created_at"`
	UpdatedAt       int64                  `json:"updated_at"`
	CompletedAt     *int64                 `json:"completed_at,omitempty"`
	ErrorMessage    *string                `json:"error_message,omitempty"`

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

	response := JobResponse{
		JobID:           job.JobID,
		Status:          job.Status,
		Stage:           job.Stage,
		Metadata:        job.Metadata,
		ProgressPercent: calculateProgress(job.Stage),
		Prompt:          job.Prompt,
		Duration:        job.Duration,
		VideoURL:        videoURL,
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       job.UpdatedAt,
		CompletedAt:     job.CompletedAt,
		ErrorMessage:    job.ErrorMessage,
		ThumbnailURL:    job.ThumbnailURL,
		AudioURL:        job.AudioURL,
		ScenesCompleted: job.ScenesCompleted,
		SceneVideoURLs:  job.SceneVideoURLs,
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

		jobResponses[i] = JobResponse{
			JobID:           job.JobID,
			Status:          job.Status,
			Stage:           job.Stage,
			Metadata:        job.Metadata,
			ProgressPercent: calculateProgress(job.Stage),
			VideoURL:        videoURL,
			ErrorMessage:    job.ErrorMessage,
			Prompt:          job.Prompt,
			Duration:        job.Duration,
			CreatedAt:       job.CreatedAt,
			UpdatedAt:       job.UpdatedAt,
			CompletedAt:     job.CompletedAt,
			ThumbnailURL:    job.ThumbnailURL,
			AudioURL:        job.AudioURL,
			ScenesCompleted: job.ScenesCompleted,
			SceneVideoURLs:  job.SceneVideoURLs,
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
