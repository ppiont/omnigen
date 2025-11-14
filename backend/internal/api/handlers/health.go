package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/repository"
	"go.uber.org/zap"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	jobRepo   *repository.DynamoDBRepository
	s3Service *repository.S3Service
	logger    *zap.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	jobRepo *repository.DynamoDBRepository,
	s3Service *repository.S3Service,
	logger *zap.Logger,
) *HealthHandler {
	return &HealthHandler{
		jobRepo:   jobRepo,
		s3Service: s3Service,
		logger:    logger,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp int64             `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// Check handles GET /health
// @Summary Health check
// @Description Check if the API is healthy and can connect to dependencies
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]string)

	// Check DynamoDB connectivity
	if err := h.jobRepo.HealthCheck(ctx); err != nil {
		h.logger.Error("DynamoDB health check failed", zap.Error(err))
		checks["dynamodb"] = "unhealthy"
	} else {
		checks["dynamodb"] = "ok"
	}

	// Check S3 connectivity
	if err := h.s3Service.HealthCheck(ctx); err != nil {
		h.logger.Error("S3 health check failed", zap.Error(err))
		checks["s3"] = "unhealthy"
	} else {
		checks["s3"] = "ok"
	}

	// Determine overall status
	status := "healthy"
	statusCode := http.StatusOK
	for _, checkStatus := range checks {
		if checkStatus != "ok" {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().Unix(),
		Checks:    checks,
	}

	c.JSON(statusCode, response)
}
