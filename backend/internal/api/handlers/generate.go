package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/internal/service"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// GenerateHandler handles video generation requests
type GenerateHandler struct {
	stepFunctions *service.StepFunctionsService
	jobRepo       *repository.DynamoDBRepository
	logger        *zap.Logger
}

// NewGenerateHandler creates a new generate handler
func NewGenerateHandler(
	stepFunctions *service.StepFunctionsService,
	jobRepo *repository.DynamoDBRepository,
	logger *zap.Logger,
) *GenerateHandler {
	return &GenerateHandler{
		stepFunctions: stepFunctions,
		jobRepo:       jobRepo,
		logger:        logger,
	}
}

// GenerateRequest represents a simplified video generation request
type GenerateRequest struct {
	Prompt      string `json:"prompt" binding:"required,min=10,max=2000"`
	Duration    int    `json:"duration" binding:"required,min=10,max=60"`
	AspectRatio string `json:"aspect_ratio" binding:"required,oneof=16:9 9:16 1:1"`
	StartImage  string `json:"start_image,omitempty" binding:"omitempty,url"`
	MusicMood   string `json:"music_mood,omitempty" binding:"omitempty,oneof=upbeat calm dramatic energetic"`
	MusicStyle  string `json:"music_style,omitempty" binding:"omitempty,oneof=electronic acoustic orchestral"`
}

// GenerateResponse represents a video generation response
type GenerateResponse struct {
	JobID               string `json:"job_id"`
	Status              string `json:"status"`
	NumClips            int    `json:"num_clips"`
	CreatedAt           int64  `json:"created_at"`
	EstimatedCompletion int    `json:"estimated_completion_seconds"`
}

// Generate handles POST /api/v1/generate
// @Summary Generate video from prompt (simplified)
// @Description Generates video by creating multiple 5s clips and stitching them together
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body GenerateRequest true "Video generation parameters"
// @Success 202 {object} GenerateResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/generate [post]
// @Security BearerAuth
func (h *GenerateHandler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest.WithDetails(map[string]interface{}{
				"validation_error": err.Error(),
			}),
		})
		return
	}

	// Validate duration is multiple of 5 (Kling constraint)
	if req.Duration%5 != 0 {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest.WithDetails(map[string]interface{}{
				"error": "duration must be a multiple of 5 seconds (Kling v2.5 limitation)",
			}),
		})
		return
	}

	// Get user ID from auth context
	userID := auth.MustGetUserID(c)

	// Calculate number of clips needed (each clip is 5 seconds)
	numClips := req.Duration / 5

	h.logger.Info("Starting video generation",
		zap.String("user_id", userID),
		zap.String("prompt", req.Prompt),
		zap.Int("duration", req.Duration),
		zap.Int("num_clips", numClips),
		zap.String("aspect_ratio", req.AspectRatio),
	)

	// Create job record
	jobID := fmt.Sprintf("job-%s", uuid.New().String())
	job := &domain.Job{
		JobID:       jobID,
		UserID:      userID,
		Status:      domain.StatusProcessing,
		Prompt:      req.Prompt,
		Duration:    req.Duration,
		AspectRatio: req.AspectRatio,
		CreatedAt:   time.Now().Unix(),
		TTL:         time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	// Save job to database
	if err := h.jobRepo.CreateJob(c.Request.Context(), job); err != nil {
		h.logger.Error("Failed to create job", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	// Default music preferences if not provided
	musicMood := req.MusicMood
	if musicMood == "" {
		musicMood = "upbeat" // Default to upbeat
	}
	musicStyle := req.MusicStyle
	if musicStyle == "" {
		musicStyle = "electronic" // Default to electronic
	}

	// Prepare Step Functions input
	sfInput := &domain.StepFunctionsInput{
		JobID:       jobID,
		Prompt:      req.Prompt,
		Duration:    req.Duration,
		AspectRatio: req.AspectRatio,
		StartImage:  req.StartImage,
		NumClips:    numClips,
		MusicMood:   musicMood,
		MusicStyle:  musicStyle,
	}

	// Start Step Functions execution
	executionARN, err := h.stepFunctions.StartExecution(c.Request.Context(), sfInput)
	if err != nil {
		h.logger.Error("Failed to start Step Functions execution",
			zap.String("job_id", jobID),
			zap.Error(err))

		// Update job status to failed
		h.jobRepo.UpdateJobStatus(c.Request.Context(), jobID, domain.StatusFailed)

		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer.WithDetails(map[string]interface{}{
				"message": "Failed to start video generation workflow",
			}),
		})
		return
	}

	h.logger.Info("Step Functions execution started",
		zap.String("job_id", jobID),
		zap.String("execution_arn", executionARN),
		zap.Int("num_clips", numClips),
	)

	// Estimate completion time: ~60s per 5s clip + 30s composition
	estimatedSeconds := (numClips * 60) + 30

	response := GenerateResponse{
		JobID:               jobID,
		Status:              job.Status,
		NumClips:            numClips,
		CreatedAt:           job.CreatedAt,
		EstimatedCompletion: estimatedSeconds,
	}

	c.JSON(http.StatusAccepted, response)
}
