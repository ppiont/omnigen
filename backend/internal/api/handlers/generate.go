package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omnigen/backend/internal/adapters"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/internal/service"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// GenerateHandler handles video generation requests with goroutine-based async processing
type GenerateHandler struct {
	parserService  *service.ParserService
	klingAdapter   *adapters.KlingAdapter
	minimaxAdapter *adapters.MinimaxAdapter
	s3Service      *repository.S3Service
	jobRepo        *repository.DynamoDBRepository
	assetsBucket   string
	logger         *zap.Logger
}

// NewGenerateHandler creates a new generate handler
func NewGenerateHandler(
	parserService *service.ParserService,
	klingAdapter *adapters.KlingAdapter,
	minimaxAdapter *adapters.MinimaxAdapter,
	s3Service *repository.S3Service,
	jobRepo *repository.DynamoDBRepository,
	assetsBucket string,
	logger *zap.Logger,
) *GenerateHandler {
	return &GenerateHandler{
		parserService:  parserService,
		klingAdapter:   klingAdapter,
		minimaxAdapter: minimaxAdapter,
		s3Service:      s3Service,
		jobRepo:        jobRepo,
		assetsBucket:   assetsBucket,
		logger:         logger,
	}
}

// GenerateRequest represents a video generation request - SIMPLE interface
type GenerateRequest struct {
	Prompt      string `json:"prompt" binding:"required,min=10,max=2000"`
	Duration    int    `json:"duration" binding:"required,min=10,max=60"`
	AspectRatio string `json:"aspect_ratio" binding:"required,oneof=16:9 9:16 1:1"`
	StartImage  string `json:"start_image,omitempty" binding:"omitempty,url"`
}

// GenerateResponse represents a video generation response
type GenerateResponse struct {
	JobID               string `json:"job_id"`
	Status              string `json:"status"`
	NumClips            int    `json:"num_clips"`
	CreatedAt           int64  `json:"created_at"`
	EstimatedCompletion int    `json:"estimated_completion_seconds"`
}

// Generate handles POST /api/v1/generate - FULLY ASYNC (returns instantly)
// @Summary Generate video from prompt with intelligent parsing
// @Description Creates job immediately and processes video generation in background goroutine
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

	// Validate duration is multiple of 10 (Kling constraint for 10s clips)
	if req.Duration%10 != 0 {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest.WithDetails(map[string]interface{}{
				"error": "duration must be a multiple of 10 seconds (Kling v2.5 10s clip limitation)",
			}),
		})
		return
	}

	// Get user ID from auth context
	userID := auth.MustGetUserID(c)

	h.logger.Info("Starting fully async video generation",
		zap.String("user_id", userID),
		zap.String("prompt", req.Prompt),
		zap.Int("duration", req.Duration),
		zap.String("aspect_ratio", req.AspectRatio),
	)

	// Create job record IMMEDIATELY (no GPT-4o call yet - that's in the goroutine!)
	jobID := fmt.Sprintf("job-%s", uuid.New().String())
	now := time.Now().Unix()
	job := &domain.Job{
		JobID:       jobID,
		UserID:      userID,
		Status:      domain.StatusProcessing,
		Stage:       "script_generating",
		Prompt:      req.Prompt,
		Duration:    req.Duration,
		AspectRatio: req.AspectRatio,
		CreatedAt:   now,
		UpdatedAt:   now,
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

	// Launch async video generation in goroutine (includes GPT-4o + video generation)
	go h.generateVideoAsync(context.Background(), job, req)

	h.logger.Info("Job created, async generation started",
		zap.String("job_id", jobID),
		zap.String("stage", "script_generating"),
	)

	// Return immediately (<100ms response time)
	response := GenerateResponse{
		JobID:               jobID,
		Status:              job.Status,
		NumClips:            0, // Will be set after script generation
		CreatedAt:           job.CreatedAt,
		EstimatedCompletion: 300, // ~5 minutes total
	}

	c.JSON(http.StatusAccepted, response)
}
