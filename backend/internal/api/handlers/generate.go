package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/service"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// GenerateHandler handles video generation requests
type GenerateHandler struct {
	generatorService *service.GeneratorService
	mockService      *service.MockService
	logger           *zap.Logger
	mockMode         bool
}

// NewGenerateHandler creates a new generate handler
func NewGenerateHandler(
	generatorService *service.GeneratorService,
	mockService *service.MockService,
	logger *zap.Logger,
	mockMode bool,
) *GenerateHandler {
	return &GenerateHandler{
		generatorService: generatorService,
		mockService:      mockService,
		logger:           logger,
		mockMode:         mockMode,
	}
}

// GenerateRequest represents a video generation request
type GenerateRequest struct {
	Prompt        string `json:"prompt" binding:"required,min=10,max=1000"`
	Duration      int    `json:"duration" binding:"required,min=15,max=180"`
	AspectRatio   string `json:"aspect_ratio" binding:"omitempty,oneof=16:9 9:16 1:1"`
	Style         string `json:"style" binding:"omitempty,max=200"`
	StartImageURL string `json:"start_image_url" binding:"omitempty,url"`
	EnableAudio   *bool  `json:"enable_audio,omitempty"` // Optional, defaults to true if not specified
}

// GenerateResponse represents a video generation response
type GenerateResponse struct {
	JobID               string `json:"job_id"`
	Status              string `json:"status"`
	CreatedAt           int64  `json:"created_at"`
	EstimatedCompletion int    `json:"estimated_completion"` // in seconds
}

// Generate handles POST /api/v1/generate
// @Summary Submit video generation job
// @Description Create a new video generation job for an ad creative
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body GenerateRequest true "Generation parameters"
// @Success 201 {object} GenerateResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 402 {object} errors.ErrorResponse "Payment Required - Monthly quota exceeded"
// @Failure 429 {object} errors.ErrorResponse "Too Many Requests - Rate limit exceeded"
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

	// Get trace ID from context
	traceID, _ := c.Get("trace_id")

	// Get user ID from auth context (or use mock user ID in mock mode)
	var userID string
	if h.mockMode {
		userID = "mock-user-local-dev"
	} else {
		userID = auth.MustGetUserID(c)
	}

	h.logger.Info("Generating video",
		zap.String("trace_id", traceID.(string)),
		zap.String("user_id", userID),
		zap.String("prompt", req.Prompt),
		zap.Int("duration", req.Duration),
		zap.Bool("mock_mode", h.mockMode),
	)

	var job *domain.Job
	var err error

	// Use mock service in mock mode, otherwise use real generation
	if h.mockMode {
		h.logger.Info("Using mock service for video generation",
			zap.String("trace_id", traceID.(string)),
			zap.String("user_id", userID),
		)
		job = h.mockService.CreateMockJob(userID, req.Prompt, req.Duration, req.AspectRatio)
	} else {
		// Default to true if not specified
		enableAudio := true
		if req.EnableAudio != nil {
			enableAudio = *req.EnableAudio
		}

		// Create domain request
		domainReq := &domain.GenerateRequest{
			UserID:        userID,
			Prompt:        req.Prompt,
			Duration:      req.Duration,
			AspectRatio:   req.AspectRatio,
			Style:         req.Style,
			StartImageURL: req.StartImageURL,
			EnableAudio:   enableAudio,
		}

		// Generate video using real pipeline
		job, err = h.generatorService.GenerateVideo(c.Request.Context(), domainReq)
		if err != nil {
			h.logger.Error("Failed to generate video",
				zap.String("trace_id", traceID.(string)),
				zap.Error(err),
			)

			// Handle different error types
			if apiErr, ok := err.(*errors.APIError); ok {
				c.JSON(apiErr.Status, errors.ErrorResponse{Error: apiErr})
				return
			}

			c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
				Error: errors.ErrInternalServer,
			})
			return
		}
	}

	// Estimate completion time
	var estimatedSeconds int
	if h.mockMode {
		// Mock mode: 180 seconds (3 minutes) fixed time
		estimatedSeconds = 180
	} else {
		// Real mode: based on video duration
		// 15s video ~= 3min, 30s ~= 5min, 60s ~= 10min
		estimatedSeconds = req.Duration * 10
	}

	response := GenerateResponse{
		JobID:               job.JobID,
		Status:              job.Status,
		CreatedAt:           job.CreatedAt,
		EstimatedCompletion: estimatedSeconds,
	}

	c.JSON(http.StatusCreated, response)
}
