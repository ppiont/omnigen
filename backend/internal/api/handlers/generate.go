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
	parserService *service.ParserService
	stepFunctions *service.StepFunctionsService
	jobRepo       *repository.DynamoDBRepository
	mockService   *service.MockService
	logger        *zap.Logger
	mockMode      bool
}

// NewGenerateHandler creates a new generate handler
func NewGenerateHandler(
	parserService *service.ParserService,
	stepFunctions *service.StepFunctionsService,
	jobRepo *repository.DynamoDBRepository,
	mockService *service.MockService,
	logger *zap.Logger,
	mockMode bool,
) *GenerateHandler {
	return &GenerateHandler{
		parserService: parserService,
		stepFunctions: stepFunctions,
		jobRepo:       jobRepo,
		mockService:   mockService,
		logger:        logger,
		mockMode:      mockMode,
	}
}

// GenerateRequest represents a video generation request
type GenerateRequest struct {
	ScriptID string `json:"script_id" binding:"required"`
}

// GenerateResponse represents a video generation response
type GenerateResponse struct {
	JobID               string `json:"job_id"`
	Status              string `json:"status"`
	CreatedAt           int64  `json:"created_at"`
	EstimatedCompletion int    `json:"estimated_completion"` // in seconds
}

// Generate handles POST /api/v1/generate
// @Summary Start video generation from approved script
// @Description Starts Step Functions workflow to generate video from an approved script
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body GenerateRequest true "Script ID to generate video from"
// @Success 202 {object} GenerateResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 402 {object} errors.ErrorResponse "Payment Required - Monthly quota exceeded"
// @Failure 404 {object} errors.ErrorResponse "Script not found"
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

	// Get user ID from auth context (or use mock user ID in mock mode)
	var userID string
	if h.mockMode {
		userID = "mock-user-local-dev"
	} else {
		userID = auth.MustGetUserID(c)
	}

	h.logger.Info("Starting video generation",
		zap.String("user_id", userID),
		zap.String("script_id", req.ScriptID),
		zap.Bool("mock_mode", h.mockMode),
	)

	// Fetch script from database
	script, err := h.parserService.GetScript(c.Request.Context(), req.ScriptID)
	if err != nil {
		h.logger.Error("Failed to fetch script",
			zap.String("script_id", req.ScriptID),
			zap.Error(err))

		// Check if it's a not found error
		if apiErr, ok := err.(*errors.APIError); ok && apiErr.Status == http.StatusNotFound {
			c.JSON(http.StatusNotFound, errors.ErrorResponse{
				Error: errors.ErrNotFound.WithDetails(map[string]interface{}{
					"script_id": req.ScriptID,
					"message":   "Script not found",
				}),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	// Verify user owns the script
	if script.UserID != userID {
		h.logger.Warn("User attempted to generate video from script they don't own",
			zap.String("user_id", userID),
			zap.String("script_id", req.ScriptID),
			zap.String("script_owner", script.UserID))

		c.JSON(http.StatusForbidden, errors.ErrorResponse{
			Error: errors.ErrForbidden.WithDetails(map[string]interface{}{
				"message": "You don't have permission to generate video from this script",
			}),
		})
		return
	}

	// Create job record
	jobID := fmt.Sprintf("job-%s", uuid.New().String())
	job := &domain.Job{
		JobID:     jobID,
		UserID:    userID,
		ScriptID:  req.ScriptID,
		Status:    domain.StatusProcessing,
		CreatedAt: time.Now().Unix(),
		TTL:       time.Now().Add(7 * 24 * time.Hour).Unix(), // 7-day TTL
	}

	// Save job to database
	if err := h.jobRepo.CreateJob(c.Request.Context(), job); err != nil {
		h.logger.Error("Failed to create job", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	var estimatedSeconds int

	// In mock mode, just create the job without starting Step Functions
	if h.mockMode {
		h.logger.Info("Mock mode: Created job without starting Step Functions",
			zap.String("job_id", jobID))
		estimatedSeconds = 180 // Mock: 3 minutes
	} else {
		// Prepare Step Functions input
		sfInput := &domain.StepFunctionsInput{
			JobID:  jobID,
			Script: script,
		}

		// Start Step Functions execution
		executionARN, err := h.stepFunctions.StartExecution(c.Request.Context(), sfInput)
		if err != nil {
			h.logger.Error("Failed to start Step Functions execution",
				zap.String("job_id", jobID),
				zap.Error(err))

			// Update job status to failed
			job.Status = domain.StatusFailed
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
			zap.String("script_id", req.ScriptID),
			zap.String("execution_arn", executionARN))

		// Real mode: based on video duration
		// 15s video ~= 3min, 30s ~= 5min, 60s ~= 10min
		estimatedSeconds = script.TotalDuration * 10
	}

	response := GenerateResponse{
		JobID:               jobID,
		Status:              job.Status,
		CreatedAt:           job.CreatedAt,
		EstimatedCompletion: estimatedSeconds,
	}

	c.JSON(http.StatusAccepted, response)
}
