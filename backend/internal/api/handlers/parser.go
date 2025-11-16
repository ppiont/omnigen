package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/service"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// ParserHandler handles script generation and management
type ParserHandler struct {
	parserService   *service.ParserService
	lambdaClient    interface{} // interface{} to avoid import cycles
	lambdaParserARN string
	logger          *zap.Logger
	mockMode        bool
}

// NewParserHandler creates a new parser handler
func NewParserHandler(
	parserService *service.ParserService,
	lambdaClient interface{},
	lambdaParserARN string,
	logger *zap.Logger,
	mockMode bool,
) *ParserHandler {
	return &ParserHandler{
		parserService:   parserService,
		lambdaClient:    lambdaClient,
		lambdaParserARN: lambdaParserARN,
		logger:          logger,
		mockMode:        mockMode,
	}
}

// ParseRequest represents a script generation request
type ParseRequest struct {
	Prompt         string `json:"prompt" binding:"required,min=10,max=1000"`
	Duration       int    `json:"duration" binding:"required,oneof=15 30 60"`
	ProductName    string `json:"product_name" binding:"required,min=2,max=100"`
	TargetAudience string `json:"target_audience" binding:"required,min=2,max=200"`
	BrandVibe      string `json:"brand_vibe,omitempty" binding:"omitempty,max=200"`
}

// ParseResponse represents a script generation response
type ParseResponse struct {
	ScriptID string `json:"script_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// Parse handles POST /api/v1/parse
// @Summary Generate ad script from user input (async)
// @Description Start async script generation using Lambda. Returns immediately with script_id for polling.
// @Tags scripts
// @Accept json
// @Produce json
// @Param request body ParseRequest true "Script generation parameters"
// @Success 202 {object} ParseResponse "Script generation started"
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/parse [post]
// @Security BearerAuth
func (h *ParserHandler) Parse(c *gin.Context) {
	var req ParseRequest
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

	h.logger.Info("Starting async script generation",
		zap.String("trace_id", traceID.(string)),
		zap.String("user_id", userID),
		zap.String("product", req.ProductName),
		zap.Int("duration", req.Duration),
	)

	// Generate script ID
	scriptID := fmt.Sprintf("script-%s", uuid.New().String())

	// Create script stub with status="generating"
	scriptStub := &domain.Script{
		ScriptID:  scriptID,
		UserID:    userID,
		Status:    "generating",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(), // 30-day TTL
	}

	// Save script stub to DynamoDB immediately
	if err := h.parserService.SaveScript(c.Request.Context(), scriptStub); err != nil {
		h.logger.Error("Failed to save script stub",
			zap.String("trace_id", traceID.(string)),
			zap.String("script_id", scriptID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer.WithDetails(map[string]interface{}{
				"error": "Failed to create script",
			}),
		})
		return
	}

	// In mock mode, call the service directly instead of Lambda
	if h.mockMode {
		h.logger.Info("Mock mode: Using synchronous script generation",
			zap.String("trace_id", traceID.(string)),
			zap.String("script_id", scriptID),
		)

		// Create service request
		serviceReq := service.ParseRequest{
			UserID:         userID,
			Prompt:         req.Prompt,
			Duration:       req.Duration,
			ProductName:    req.ProductName,
			TargetAudience: req.TargetAudience,
			BrandVibe:      req.BrandVibe,
		}

		// Generate script synchronously in mock mode
		script, err := h.parserService.GenerateScript(c.Request.Context(), serviceReq)
		if err != nil {
			h.logger.Error("Failed to generate script",
				zap.String("trace_id", traceID.(string)),
				zap.Error(err),
			)
			// Update status to failed
			scriptStub.Status = "failed"
			h.parserService.SaveScript(c.Request.Context(), scriptStub)

			c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
				Error: errors.ErrInternalServer,
			})
			return
		}

		h.logger.Info("Mock mode: Script generated successfully",
			zap.String("trace_id", traceID.(string)),
			zap.String("script_id", script.ScriptID),
		)

		response := ParseResponse{
			ScriptID: scriptID,
			Status:   "draft",
			Message:  "Script generated successfully (mock mode)",
		}

		c.JSON(http.StatusAccepted, response)
		return
	}

	// Prepare Lambda invocation payload
	lambdaInput := map[string]interface{}{
		"script_id":       scriptID,
		"user_id":         userID,
		"prompt":          req.Prompt,
		"duration":        req.Duration,
		"product_name":    req.ProductName,
		"target_audience": req.TargetAudience,
	}

	if req.BrandVibe != "" {
		lambdaInput["brand_vibe"] = req.BrandVibe
	}

	payload, err := json.Marshal(lambdaInput)
	if err != nil {
		h.logger.Error("Failed to marshal Lambda payload",
			zap.String("trace_id", traceID.(string)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	// Invoke Parser Lambda asynchronously
	lambdaClient, ok := h.lambdaClient.(*lambda.Client)
	if !ok || lambdaClient == nil {
		h.logger.Error("Lambda client not configured",
			zap.String("trace_id", traceID.(string)),
			zap.String("script_id", scriptID),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer.WithDetails(map[string]interface{}{
				"error": "Lambda client not configured",
			}),
		})
		return
	}

	_, err = lambdaClient.Invoke(c.Request.Context(), &lambda.InvokeInput{
		FunctionName:   aws.String(h.lambdaParserARN),
		InvocationType: types.InvocationTypeEvent, // Async invocation
		Payload:        payload,
	})

	if err != nil {
		h.logger.Error("Failed to invoke Parser Lambda",
			zap.String("trace_id", traceID.(string)),
			zap.String("script_id", scriptID),
			zap.Error(err),
		)

		// Mark script as failed
		scriptStub.Status = "failed"
		h.parserService.SaveScript(c.Request.Context(), scriptStub)

		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer.WithDetails(map[string]interface{}{
				"error": "Failed to start script generation",
			}),
		})
		return
	}

	h.logger.Info("Parser Lambda invoked successfully",
		zap.String("trace_id", traceID.(string)),
		zap.String("script_id", scriptID),
	)

	response := ParseResponse{
		ScriptID: scriptID,
		Status:   "generating",
		Message:  "Script generation started. Use GET /scripts/{script_id} to check status.",
	}

	c.JSON(http.StatusAccepted, response)
}

// GetScript handles GET /api/v1/scripts/:id
// @Summary Retrieve a script by ID
// @Description Get a previously generated script for review or editing
// @Tags scripts
// @Produce json
// @Param id path string true "Script ID"
// @Success 200 {object} domain.Script
// @Failure 401 {object} errors.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 404 {object} errors.ErrorResponse "Script not found"
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/scripts/{id} [get]
// @Security BearerAuth
func (h *ParserHandler) GetScript(c *gin.Context) {
	scriptID := c.Param("id")
	if scriptID == "" {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest.WithDetails(map[string]interface{}{
				"error": "script_id is required",
			}),
		})
		return
	}

	// Get trace ID from context
	traceID, _ := c.Get("trace_id")

	h.logger.Info("Retrieving script",
		zap.String("trace_id", traceID.(string)),
		zap.String("script_id", scriptID),
	)

	script, err := h.parserService.GetScript(c.Request.Context(), scriptID)
	if err != nil {
		h.logger.Error("Failed to retrieve script",
			zap.String("trace_id", traceID.(string)),
			zap.String("script_id", scriptID),
			zap.Error(err),
		)

		if apiErr, ok := err.(*errors.APIError); ok {
			c.JSON(apiErr.Status, errors.ErrorResponse{Error: apiErr})
			return
		}

		// Check if script not found
		if err.Error() == "script not found: "+scriptID {
			c.JSON(http.StatusNotFound, errors.ErrorResponse{
				Error: errors.ErrNotFound.WithDetails(map[string]interface{}{
					"script_id": scriptID,
				}),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, script)
}

// UpdateScriptRequest represents a script update request
type UpdateScriptRequest struct {
	Script domain.Script `json:"script" binding:"required"`
}

// UpdateScript handles PUT /api/v1/scripts/:id
// @Summary Update an existing script
// @Description Update a script after user review/editing
// @Tags scripts
// @Accept json
// @Produce json
// @Param id path string true "Script ID"
// @Param request body UpdateScriptRequest true "Updated script"
// @Success 200 {object} domain.Script
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 404 {object} errors.ErrorResponse "Script not found"
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/scripts/{id} [put]
// @Security BearerAuth
func (h *ParserHandler) UpdateScript(c *gin.Context) {
	scriptID := c.Param("id")
	if scriptID == "" {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest.WithDetails(map[string]interface{}{
				"error": "script_id is required",
			}),
		})
		return
	}

	var req UpdateScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest.WithDetails(map[string]interface{}{
				"validation_error": err.Error(),
			}),
		})
		return
	}

	// Ensure script ID matches URL parameter
	if req.Script.ScriptID != scriptID {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest.WithDetails(map[string]interface{}{
				"error": "script_id in body must match URL parameter",
			}),
		})
		return
	}

	// Get trace ID from context
	traceID, _ := c.Get("trace_id")

	h.logger.Info("Updating script",
		zap.String("trace_id", traceID.(string)),
		zap.String("script_id", scriptID),
		zap.Int("num_scenes", len(req.Script.Scenes)),
	)

	err := h.parserService.UpdateScript(c.Request.Context(), &req.Script)
	if err != nil {
		h.logger.Error("Failed to update script",
			zap.String("trace_id", traceID.(string)),
			zap.String("script_id", scriptID),
			zap.Error(err),
		)

		if apiErr, ok := err.(*errors.APIError); ok {
			c.JSON(apiErr.Status, errors.ErrorResponse{Error: apiErr})
			return
		}

		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	h.logger.Info("Script updated successfully",
		zap.String("trace_id", traceID.(string)),
		zap.String("script_id", scriptID),
	)

	c.JSON(http.StatusOK, &req.Script)
}
