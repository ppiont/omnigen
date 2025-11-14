package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/service"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// GenerateHandler handles video generation requests
type GenerateHandler struct {
	generatorService *service.GeneratorService
	logger           *zap.Logger
}

// NewGenerateHandler creates a new generate handler
func NewGenerateHandler(
	generatorService *service.GeneratorService,
	logger *zap.Logger,
) *GenerateHandler {
	return &GenerateHandler{
		generatorService: generatorService,
		logger:           logger,
	}
}

// GenerateRequest represents a video generation request
type GenerateRequest struct {
	Prompt      string `json:"prompt" binding:"required,min=10,max=1000"`
	Duration    int    `json:"duration" binding:"required,min=15,max=180"`
	AspectRatio string `json:"aspect_ratio" binding:"omitempty,oneof=16:9 9:16 1:1"`
	Style       string `json:"style" binding:"omitempty,max=200"`
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
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/generate [post]
// @Security ApiKeyAuth
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

	h.logger.Info("Generating video",
		zap.String("trace_id", traceID.(string)),
		zap.String("prompt", req.Prompt),
		zap.Int("duration", req.Duration),
	)

	// Create domain request
	domainReq := &domain.GenerateRequest{
		Prompt:      req.Prompt,
		Duration:    req.Duration,
		AspectRatio: req.AspectRatio,
		Style:       req.Style,
	}

	// Generate video
	job, err := h.generatorService.GenerateVideo(c.Request.Context(), domainReq)
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

	// Estimate completion time (based on duration)
	// 15s video ~= 3min, 30s ~= 5min, 60s ~= 10min
	estimatedSeconds := req.Duration * 10

	response := GenerateResponse{
		JobID:               job.JobID,
		Status:              job.Status,
		CreatedAt:           job.CreatedAt,
		EstimatedCompletion: estimatedSeconds,
	}

	c.JSON(http.StatusCreated, response)
}
