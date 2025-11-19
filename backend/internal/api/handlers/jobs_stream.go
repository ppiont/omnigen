package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/domain"
	"go.uber.org/zap"
)

// JobUpdateEvent represents a real-time job update for SSE
type JobUpdateEvent struct {
	Stage    string `json:"stage"`
	Status   string `json:"status"`
	VideoKey string `json:"video_key,omitempty"`
	Progress int    `json:"progress"`
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
					VideoKey: job.VideoKey,
					Progress: calculateDynamicProgress(job.Stage, len(job.Scenes)),
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

// calculateETA estimates time remaining based on elapsed time and current progress
// Returns estimated seconds remaining
func calculateETA(stage string, startTime time.Time, totalScenes int) int {
	progress := calculateDynamicProgress(stage, totalScenes)

	// If no progress yet or completed, return 0
	if progress == 0 || progress >= 100 {
		return 0
	}

	// Calculate elapsed time in seconds
	elapsed := time.Since(startTime).Seconds()

	// Estimate total time based on current progress
	// totalEstimated = elapsed * (100 / progress)
	totalEstimated := elapsed * 100.0 / float64(progress)

	// Remaining time = total - elapsed
	remaining := totalEstimated - elapsed

	// Return 0 if negative (shouldn't happen, but safety check)
	if remaining < 0 {
		return 0
	}

	return int(remaining)
}
