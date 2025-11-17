package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/domain"
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
	}

	c.JSON(http.StatusOK, response)
}

// ListJobs handles GET /api/v1/jobs
// @Summary List jobs
// @Description Get a list of video generation jobs for the authenticated user
// @Tags jobs
// @Produce json
// @Param page_size query int false "Page size (max 100)" default(20)
// @Param status query string false "Filter by status (pending, processing, completed, failed)"
// @Param next_token query string false "Pagination token from previous response"
// @Success 200 {object} ListJobsResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/jobs [get]
// @Security BearerAuth
func (h *JobsHandler) ListJobs(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, errors.ErrorResponse{
			Error: errors.ErrUnauthorized.WithDetails(map[string]interface{}{
				"message": "Authentication required",
			}),
		})
		return
	}

	// Parse page size
	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize := 20
	if ps, err := parsePageSize(pageSizeStr); err == nil {
		pageSize = ps
	}

	// Get status filter
	statusFilter := c.Query("status")

	// Get pagination token (base64 encoded LastEvaluatedKey)
	nextToken := c.Query("next_token")

	h.logger.Info("Listing jobs",
		zap.String("user_id", userID.(string)),
		zap.Int("page_size", pageSize),
		zap.String("status", statusFilter),
		zap.Bool("has_next_token", nextToken != ""),
	)

	// Decode pagination token if present
	var lastEvaluatedKey map[string]types.AttributeValue
	if nextToken != "" {
		// For now, we'll skip token decoding - in production you'd decode base64
		// This is a simplification for the MVP
		h.logger.Debug("Pagination token provided but not yet implemented")
	}

	// Query jobs from DynamoDB
	jobs, newLastEvaluatedKey, err := h.jobRepo.ListJobsByUser(
		c.Request.Context(),
		userID.(string),
		pageSize,
		lastEvaluatedKey,
	)
	if err != nil {
		h.logger.Error("Failed to list jobs",
			zap.String("user_id", userID.(string)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrDatabaseError,
		})
		return
	}

	// Apply status filter if provided
	var filteredJobs []*domain.Job
	for i := range jobs {
		if statusFilter == "" || jobs[i].Status == statusFilter {
			filteredJobs = append(filteredJobs, &jobs[i])
		}
	}

	// Convert to response format
	jobResponses := make([]JobResponse, 0, len(filteredJobs))
	for _, job := range filteredJobs {
		// Generate presigned URL if video is completed
		var videoURL *string
		if job.Status == "completed" && job.VideoKey != "" {
			url, err := h.s3Service.GetPresignedURL(c.Request.Context(), job.VideoKey, 7*24*time.Hour)
			if err != nil {
				h.logger.Error("Failed to generate presigned URL",
					zap.String("job_id", job.JobID),
					zap.Error(err),
				)
			} else {
				videoURL = &url
			}
		}

		jobResponses = append(jobResponses, JobResponse{
			JobID:        job.JobID,
			Status:       job.Status,
			Prompt:       job.Prompt,
			Duration:     job.Duration,
			Style:        job.Style,
			VideoURL:     videoURL,
			CreatedAt:    job.CreatedAt,
			CompletedAt:  job.CompletedAt,
			ErrorMessage: job.ErrorMessage,
		})
	}

	// Build response
	response := ListJobsResponse{
		Jobs:       jobResponses,
		TotalCount: len(jobResponses),
		Page:       1, // Simplified for now - token-based pagination doesn't use pages
		PageSize:   pageSize,
	}

	// Add next token if there are more results
	if newLastEvaluatedKey != nil {
		// In production, encode this as base64
		// For now, we'll just indicate there are more results
		h.logger.Debug("More results available",
			zap.Int("returned_count", len(jobResponses)),
		)
	}

	c.JSON(http.StatusOK, response)
}

// DeleteJob handles DELETE /api/v1/jobs/:id
// @Summary Delete a job
// @Description Delete a video generation job and its associated files
// @Tags jobs
// @Produce json
// @Param id path string true "Job ID"
// @Success 204 "Job deleted successfully"
// @Failure 404 {object} errors.ErrorResponse "Job not found"
// @Failure 403 {object} errors.ErrorResponse "Forbidden - not the job owner"
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/jobs/{id} [delete]
// @Security BearerAuth
func (h *JobsHandler) DeleteJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest.WithDetails(map[string]interface{}{
				"error": "job_id is required",
			}),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, errors.ErrorResponse{
			Error: errors.ErrUnauthorized.WithDetails(map[string]interface{}{
				"message": "Authentication required",
			}),
		})
		return
	}

	h.logger.Info("Deleting job",
		zap.String("job_id", jobID),
		zap.String("user_id", userID.(string)),
	)

	// Get the job first to verify ownership and get video key
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

	// Verify the user owns this job
	if job.UserID != userID.(string) {
		h.logger.Warn("User attempted to delete job they don't own",
			zap.String("job_id", jobID),
			zap.String("user_id", userID.(string)),
			zap.String("job_owner", job.UserID),
		)
		c.JSON(http.StatusForbidden, errors.ErrorResponse{
			Error: errors.ErrForbidden.WithDetails(map[string]interface{}{
				"message": "You don't have permission to delete this job",
			}),
		})
		return
	}

	// Delete the video file from S3 if it exists
	if job.VideoKey != "" {
		h.logger.Info("Deleting video file from S3",
			zap.String("job_id", jobID),
			zap.String("video_key", job.VideoKey),
		)

		if err := h.s3Service.DeleteObject(c.Request.Context(), job.VideoKey); err != nil {
			// Log the error but continue with job deletion
			h.logger.Error("Failed to delete video from S3 (continuing with job deletion)",
				zap.String("job_id", jobID),
				zap.String("video_key", job.VideoKey),
				zap.Error(err),
			)
		}
	}

	// Delete the job from DynamoDB
	if err := h.jobRepo.DeleteJob(c.Request.Context(), jobID); err != nil {
		h.logger.Error("Failed to delete job",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrDatabaseError,
		})
		return
	}

	h.logger.Info("Job deleted successfully",
		zap.String("job_id", jobID),
		zap.String("user_id", userID.(string)),
	)

	// Return 204 No Content on successful deletion
	c.Status(http.StatusNoContent)
}

// parsePageSize parses and validates page size
func parsePageSize(s string) (int, error) {
	var size int
	if _, err := fmt.Sscanf(s, "%d", &size); err != nil {
		return 0, err
	}
	if size < 1 {
		size = 1
	}
	if size > 100 {
		size = 100
	}
	return size, nil
}
