package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/service"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// ParserHandler handles script generation and management
type ParserHandler struct {
	parserService *service.ParserService
	logger        *zap.Logger
}

// NewParserHandler creates a new parser handler
func NewParserHandler(
	parserService *service.ParserService,
	logger *zap.Logger,
) *ParserHandler {
	return &ParserHandler{
		parserService: parserService,
		logger:        logger,
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

	// Get user ID from auth context
	userID := auth.MustGetUserID(c)

	h.logger.Info("Starting async script generation",
		zap.String("trace_id", traceID.(string)),
		zap.String("user_id", userID),
		zap.String("product", req.ProductName),
		zap.Int("duration", req.Duration),
	)

	// Build comprehensive prompt from request fields
	fullPrompt := fmt.Sprintf("%s. Product: %s. Target audience: %s", req.Prompt, req.ProductName, req.TargetAudience)
	if req.BrandVibe != "" {
		fullPrompt += fmt.Sprintf(". Brand vibe: %s", req.BrandVibe)
	}

	// Call ParserService directly to generate script
	script, err := h.parserService.GenerateScript(c.Request.Context(), service.ParseRequest{
		UserID:      userID,
		Prompt:      fullPrompt,
		Duration:    req.Duration,
		AspectRatio: "16:9", // Default aspect ratio for /parse endpoint
	})

	if err != nil {
		h.logger.Error("Failed to generate script",
			zap.String("trace_id", traceID.(string)),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer.WithDetails(map[string]interface{}{
				"error": "Failed to generate script",
			}),
		})
		return
	}

	h.logger.Info("Script generated successfully",
		zap.String("trace_id", traceID.(string)),
		zap.String("script_id", script.ScriptID),
	)

	response := ParseResponse{
		ScriptID: script.ScriptID,
		Status:   script.Status,
		Message:  "Script generated successfully.",
	}

	c.JSON(http.StatusOK, response)
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
