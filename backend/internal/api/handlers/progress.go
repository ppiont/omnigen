package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// ProgressHandler handles job progress requests
type ProgressHandler struct {
	logger *zap.Logger
}

// NewProgressHandler creates a new progress handler
func NewProgressHandler(
	logger *zap.Logger,
) *ProgressHandler {
	return &ProgressHandler{
		logger: logger,
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

	// Progress tracking not yet implemented
	h.logger.Warn("Progress endpoint called but not yet implemented",
		zap.String("job_id", jobID),
	)
	c.JSON(http.StatusNotImplemented, errors.ErrorResponse{
		Error: errors.ErrNotImplemented.WithDetails(map[string]interface{}{
			"message": "Progress tracking not yet implemented. Use GET /api/v1/jobs/:id instead",
		}),
	})
}

// formatStageName converts internal stage names to user-friendly descriptions
func formatStageName(stage string) string {
	stageNames := map[string]string{
		"pending":           "Queued for processing",
		"parsing":           "Analyzing your prompt",
		"generating_videos": "Generating video clips",
		"generating_audio":  "Creating background music and voiceover",
		"composing":         "Composing final video",
		"completed":         "Video ready",
		"failed":            "Generation failed",
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
