package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/service"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// ProgressHandler handles job progress requests
type ProgressHandler struct {
	mockService *service.MockService
	logger      *zap.Logger
	mockMode    bool
}

// NewProgressHandler creates a new progress handler
func NewProgressHandler(
	mockService *service.MockService,
	logger *zap.Logger,
	mockMode bool,
) *ProgressHandler {
	return &ProgressHandler{
		mockService: mockService,
		logger:      logger,
		mockMode:    mockMode,
	}
}

// ProgressResponse represents a detailed progress response
type ProgressResponse struct {
	JobID                  string   `json:"job_id"`
	Status                 string   `json:"status"`
	Progress               int      `json:"progress"`
	CurrentStage           string   `json:"current_stage"`
	StagesCompleted        []string `json:"stages_completed"`
	StagesPending          []string `json:"stages_pending"`
	EstimatedTimeRemaining int      `json:"estimated_time_remaining"` // seconds
}

// GetProgress handles GET /api/v1/jobs/:id/progress
// @Summary Get detailed job progress
// @Description Get real-time progress updates for a video generation job
// @Tags jobs
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} ProgressResponse
// @Failure 404 {object} errors.ErrorResponse "Job not found"
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/jobs/{id}/progress [get]
// @Security BearerAuth
func (h *ProgressHandler) GetProgress(c *gin.Context) {
	jobID := c.Param("id")

	if !h.mockMode {
		// In production mode, return error as this endpoint is not yet implemented
		h.logger.Warn("Progress endpoint called in non-mock mode",
			zap.String("job_id", jobID),
		)
		c.JSON(http.StatusNotImplemented, errors.ErrorResponse{
			Error: errors.ErrNotImplemented.WithDetails(map[string]interface{}{
				"message": "Progress tracking not yet implemented. Use GET /api/v1/jobs/:id instead",
			}),
		})
		return
	}

	// Mock mode: return realistic progress
	progress := h.mockService.GetMockProgress(jobID)
	if progress == nil {
		h.logger.Warn("Mock job not found",
			zap.String("job_id", jobID),
		)
		c.JSON(http.StatusNotFound, errors.ErrorResponse{
			Error: errors.ErrNotFound.WithDetails(map[string]interface{}{
				"resource": "job",
				"job_id":   jobID,
			}),
		})
		return
	}

	response := ProgressResponse{
		JobID:                  progress.JobID,
		Status:                 progress.Status,
		Progress:               progress.Progress,
		CurrentStage:           formatStageName(progress.CurrentStage),
		StagesCompleted:        formatStageNames(progress.StagesCompleted),
		StagesPending:          formatStageNames(progress.StagesPending),
		EstimatedTimeRemaining: progress.EstimatedTimeRemaining,
	}

	h.logger.Info("Progress retrieved",
		zap.String("job_id", jobID),
		zap.String("status", progress.Status),
		zap.Int("progress", progress.Progress),
	)

	c.JSON(http.StatusOK, response)
}

// formatStageName converts internal stage names to user-friendly descriptions
func formatStageName(stage string) string {
	stageNames := map[string]string{
		"pending":            "Queued for processing",
		"parsing":            "Analyzing your prompt",
		"generating_videos":  "Generating video clips",
		"generating_audio":   "Creating background music and voiceover",
		"composing":          "Composing final video",
		"completed":          "Video ready",
		"failed":             "Generation failed",
	}

	if name, ok := stageNames[stage]; ok {
		return name
	}
	return stage
}

// formatStageNames converts a slice of internal stage names to user-friendly descriptions
func formatStageNames(stages []string) []string {
	formatted := make([]string, len(stages))
	for i, stage := range stages {
		formatted[i] = formatStageName(stage)
	}
	return formatted
}
