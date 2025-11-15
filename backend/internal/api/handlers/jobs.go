package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/internal/service"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// JobsHandler handles job-related requests
type JobsHandler struct {
	jobRepo     *repository.DynamoDBRepository
	s3Service   *repository.S3Service
	mockService *service.MockService
	logger      *zap.Logger
	mockMode    bool
}

// NewJobsHandler creates a new jobs handler
func NewJobsHandler(
	jobRepo *repository.DynamoDBRepository,
	s3Service *repository.S3Service,
	mockService *service.MockService,
	logger *zap.Logger,
	mockMode bool,
) *JobsHandler {
	return &JobsHandler{
		jobRepo:     jobRepo,
		s3Service:   s3Service,
		mockService: mockService,
		logger:      logger,
		mockMode:    mockMode,
	}
}

// JobResponse represents a job status response
type JobResponse struct {
	JobID        string  `json:"job_id"`
	Status       string  `json:"status"`
	Prompt       string  `json:"prompt"`
	Duration     int     `json:"duration"`
	Style        string  `json:"style,omitempty"`
	VideoURL     *string `json:"video_url,omitempty"`
	CreatedAt    int64   `json:"created_at"`
	CompletedAt  *int64  `json:"completed_at,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
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

	var job *domain.Job
	var err error

	// Use mock service in mock mode
	if h.mockMode {
		h.logger.Info("Getting job from mock service",
			zap.String("job_id", jobID),
		)
		job = h.mockService.GetMockJobStatus(jobID)
		if job == nil {
			c.JSON(http.StatusNotFound, errors.ErrorResponse{
				Error: errors.ErrJobNotFound,
			})
			return
		}
	} else {
		// Get job from DynamoDB in real mode
		job, err = h.jobRepo.GetJob(c.Request.Context(), jobID)
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
	}

	// Generate presigned URL if video is completed
	var videoURL *string
	if job.Status == "completed" && job.VideoKey != "" {
		if h.mockMode {
			// In mock mode, create a mock S3 URL
			mockURL := fmt.Sprintf("https://omnigen-assets-local.s3.amazonaws.com/%s", job.VideoKey)
			videoURL = &mockURL
		} else {
			// In real mode, generate presigned URL
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
	}

	response := JobResponse{
		JobID:        job.JobID,
		Status:       job.Status,
		Prompt:       job.Prompt,
		Duration:     job.Duration,
		Style:        job.Style,
		VideoURL:     videoURL,
		CreatedAt:    job.CreatedAt,
		CompletedAt:  job.CompletedAt,
		ErrorMessage: job.ErrorMessage,
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
	// Get query parameters
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")
	status := c.Query("status")

	h.logger.Info("Listing jobs",
		zap.String("page", page),
		zap.String("page_size", pageSize),
		zap.String("status", status),
	)

	// For MVP, return a simple response
	// TODO: Implement pagination and filtering
	response := ListJobsResponse{
		Jobs:       []JobResponse{},
		TotalCount: 0,
		Page:       1,
		PageSize:   20,
	}

	c.JSON(http.StatusOK, response)
}
