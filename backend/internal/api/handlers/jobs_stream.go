package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/domain"
	"go.uber.org/zap"
)

// JobUpdateEvent represents a real-time job update for SSE
type JobUpdateEvent struct {
	Stage    string                 `json:"stage"`
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	VideoKey string                 `json:"video_key,omitempty"`
	Progress int                    `json:"progress"`
}

// StreamJobUpdates handles GET /api/v1/jobs/:id/stream - Server-Sent Events endpoint
// @Summary Stream job status updates in real-time
// @Description Streams job progress updates using Server-Sent Events (SSE)
// @Tags jobs
// @Produce text/event-stream
// @Param id path string true "Job ID"
// @Success 200 {string} string "Event stream"
// @Failure 404 {object} errors.ErrorResponse
// @Router /api/v1/jobs/{id}/stream [get]
// @Security BearerAuth
func (h *JobsHandler) StreamJobUpdates(c *gin.Context) {
	jobID := c.Param("id")

	h.logger.Info("Starting SSE stream for job", zap.String("job_id", jobID))

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create ticker for polling DynamoDB
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastStage := ""
	ctx := c.Request.Context()

	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			h.logger.Info("SSE client disconnected", zap.String("job_id", jobID))
			return

		case <-ticker.C:
			// Poll DynamoDB for job status
			job, err := h.jobRepo.GetJob(context.Background(), jobID)
			if err != nil {
				h.logger.Error("Failed to get job in SSE stream",
					zap.String("job_id", jobID),
					zap.Error(err),
				)
				// Send error event
				c.SSEvent("error", "Failed to fetch job status")
				c.Writer.Flush()
				continue
			}

			// Only send update if stage changed
			if job.Stage != lastStage {
				lastStage = job.Stage

				event := JobUpdateEvent{
					Stage:    job.Stage,
					Status:   job.Status,
					Metadata: job.Metadata,
					VideoKey: job.VideoKey,
					Progress: calculateProgress(job.Stage),
				}

				data, err := json.Marshal(event)
				if err != nil {
					h.logger.Error("Failed to marshal event",
						zap.String("job_id", jobID),
						zap.Error(err),
					)
					continue
				}

				h.logger.Debug("Sending SSE update",
					zap.String("job_id", jobID),
					zap.String("stage", job.Stage),
					zap.Int("progress", event.Progress),
				)

				c.SSEvent("update", string(data))
				c.Writer.Flush()
			}

			// Close stream when job complete or failed
			if job.Status == domain.StatusCompleted || job.Status == domain.StatusFailed {
				h.logger.Info("Job terminal state reached, closing SSE stream",
					zap.String("job_id", jobID),
					zap.String("status", job.Status),
				)

				// Send final complete event
				c.SSEvent("done", fmt.Sprintf(`{"status": "%s"}`, job.Status))
				c.Writer.Flush()
				return
			}
		}
	}
}

// calculateProgress returns progress percentage based on stage name
func calculateProgress(stage string) int {
	switch {
	case stage == "script_generating":
		return 5
	case stage == "script_complete":
		return 15
	case strings.HasPrefix(stage, "scene_1_generating"):
		return 20
	case strings.HasPrefix(stage, "scene_1_complete"):
		return 30
	case strings.HasPrefix(stage, "scene_2_generating"):
		return 40
	case strings.HasPrefix(stage, "scene_2_complete"):
		return 50
	case strings.HasPrefix(stage, "scene_3_generating"):
		return 60
	case strings.HasPrefix(stage, "scene_3_complete"):
		return 70
	case strings.HasPrefix(stage, "scene_4_generating"):
		return 75
	case strings.HasPrefix(stage, "scene_4_complete"):
		return 78
	case strings.HasPrefix(stage, "scene_5_generating"):
		return 80
	case strings.HasPrefix(stage, "scene_5_complete"):
		return 82
	case strings.HasPrefix(stage, "scene_6_generating"):
		return 84
	case strings.HasPrefix(stage, "scene_6_complete"):
		return 86
	case stage == "audio_generating":
		return 88
	case stage == "audio_complete":
		return 92
	case stage == "composing":
		return 95
	case stage == "complete":
		return 100
	case stage == "failed":
		return 0
	default:
		return 0
	}
}
