package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/internal/service"
	"go.uber.org/zap"
)

// ProgressHandler handles job progress requests
type ProgressHandler struct {
	jobRepo      repository.JobRepository
	assetService *service.AssetService
	logger       *zap.Logger
}

// NewProgressHandler creates a new progress handler
func NewProgressHandler(
	jobRepo repository.JobRepository,
	assetService *service.AssetService,
	logger *zap.Logger,
) *ProgressHandler {
	return &ProgressHandler{
		jobRepo:      jobRepo,
		assetService: assetService,
		logger:       logger,
	}
}

// ProgressResponse contains comprehensive job progress information
type ProgressResponse struct {
	JobID                  string          `json:"job_id"`
	Status                 string          `json:"status"`
	Progress               int             `json:"progress"`
	CurrentStage           string          `json:"current_stage"`
	CurrentStageDisplay    string          `json:"current_stage_display"`
	StagesCompleted        []StageInfo     `json:"stages_completed"`
	StagesPending          []StageInfo     `json:"stages_pending"`
	EstimatedTimeRemaining int             `json:"estimated_time_remaining"`
	Assets                 *ProgressAssets `json:"assets,omitempty"`
}

// StageInfo contains information about a pipeline stage
type StageInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Progress    int    `json:"progress"`
	CompletedAt *int64 `json:"completed_at,omitempty"`
}

// ProgressAssets contains presigned URLs for all job assets
type ProgressAssets struct {
	SceneClips []AssetInfo `json:"scene_clips"`
	Thumbnails []AssetInfo `json:"thumbnails"`
	Audio      *AssetInfo  `json:"audio,omitempty"`
	FinalVideo *AssetInfo  `json:"final_video,omitempty"`
}

// AssetInfo contains metadata about a single asset
type AssetInfo struct {
	URL         string `json:"url"`
	SceneNumber int    `json:"scene_number,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

// GetProgress handles GET /api/v1/jobs/:id/progress
// @Summary Stream job progress in real-time
// @Description Streams comprehensive job progress updates using Server-Sent Events (SSE). Sends ProgressResponse objects as SSE events whenever the job stage changes. Automatically closes stream when job completes or fails.
// @Tags jobs
// @Produce text/event-stream
// @Param id path string true "Job ID"
// @Success 200 {string} string "Event stream (event types: update, done, error)"
// @Router /api/v1/jobs/{id}/progress [get]
// @Security BearerAuth
func (h *ProgressHandler) GetProgress(c *gin.Context) {
	jobID := c.Param("id")

	h.logger.Info("Starting SSE progress stream for job",
		zap.String("job_id", jobID),
	)

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
			h.logger.Info("SSE client disconnected",
				zap.String("job_id", jobID),
			)
			return

		case <-ticker.C:
			// Fetch job from DynamoDB
			job, err := h.jobRepo.GetJob(context.Background(), jobID)
			if err != nil {
				h.logger.Error("Failed to get job in SSE stream",
					zap.String("job_id", jobID),
					zap.Error(err),
				)
				// Send error event
				c.SSEvent("error", gin.H{"error": "Failed to fetch job status"})
				c.Writer.Flush()
				continue
			}

			// Only send update if stage changed (avoid spam)
			if job.Stage != lastStage {
				lastStage = job.Stage

				// Build full progress response
				response, err := h.buildProgressResponse(job)
				if err != nil {
					h.logger.Error("Failed to build progress response",
						zap.String("job_id", jobID),
						zap.Error(err),
					)
					continue
				}

				// Marshal to JSON
				data, err := json.Marshal(response)
				if err != nil {
					h.logger.Error("Failed to marshal progress response",
						zap.String("job_id", jobID),
						zap.Error(err),
					)
					continue
				}

				h.logger.Debug("Sending SSE progress update",
					zap.String("job_id", jobID),
					zap.String("stage", job.Stage),
					zap.Int("progress", response.Progress),
				)

				// Send update event
				c.SSEvent("update", string(data))
				c.Writer.Flush()
			}

			// Close stream when job reaches terminal state
			if job.Status == domain.StatusCompleted || job.Status == domain.StatusFailed {
				h.logger.Info("Job terminal state reached, closing SSE stream",
					zap.String("job_id", jobID),
					zap.String("status", job.Status),
				)

				// Send final done event
				c.SSEvent("done", gin.H{"status": job.Status})
				c.Writer.Flush()
				return
			}
		}
	}
}

// buildProgressResponse constructs a complete ProgressResponse from a job
func (h *ProgressHandler) buildProgressResponse(job *domain.Job) (*ProgressResponse, error) {
	// Calculate progress percentage and ETA
	progress := calculateDynamicProgress(job.Stage, len(job.Scenes))
	eta := calculateETA(job.Stage, time.Unix(job.CreatedAt, 0), len(job.Scenes))

	// Generate presigned URLs for all assets
	assets, err := h.assetService.GetJobAssets(context.Background(), job, 1*time.Hour)
	if err != nil {
		h.logger.Warn("Failed to generate asset URLs",
			zap.String("job_id", job.JobID),
			zap.Error(err),
		)
		// Continue without assets rather than failing entire request
		assets = &service.JobAssets{}
	}

	// Convert JobAssets to ProgressAssets format
	progressAssets := &ProgressAssets{
		SceneClips: make([]AssetInfo, 0, len(assets.SceneClips)),
		Thumbnails: make([]AssetInfo, 0, len(assets.Thumbnails)),
	}

	// Add scene clips
	for i, url := range assets.SceneClips {
		progressAssets.SceneClips = append(progressAssets.SceneClips, AssetInfo{
			URL:         url,
			SceneNumber: i + 1,
			CreatedAt:   job.UpdatedAt,
		})
	}

	// Add thumbnails
	for i, url := range assets.Thumbnails {
		progressAssets.Thumbnails = append(progressAssets.Thumbnails, AssetInfo{
			URL:         url,
			SceneNumber: i + 1,
			CreatedAt:   job.UpdatedAt,
		})
	}

	// Add audio
	if assets.AudioURL != "" {
		progressAssets.Audio = &AssetInfo{
			URL:       assets.AudioURL,
			CreatedAt: job.UpdatedAt,
		}
	}

	// Add final video
	if assets.FinalVideoURL != "" {
		progressAssets.FinalVideo = &AssetInfo{
			URL:       assets.FinalVideoURL,
			CreatedAt: *job.CompletedAt,
		}
	}

	// Build response
	response := &ProgressResponse{
		JobID:                  job.JobID,
		Status:                 job.Status,
		Progress:               progress,
		CurrentStage:           job.Stage,
		CurrentStageDisplay:    formatStageName(job.Stage),
		StagesCompleted:        buildStagesCompleted(job),
		StagesPending:          buildStagesPending(job),
		EstimatedTimeRemaining: eta,
		Assets:                 progressAssets,
	}

	return response, nil
}

// formatStageName converts internal stage names to user-friendly display names
func formatStageName(stage string) string {
	switch stage {
	case "script_generating":
		return "Generating script with AI"
	case "script_complete":
		return "Script ready"
	case "audio_generating":
		return "Generating background music"
	case "audio_complete":
		return "Audio ready"
	case "composing":
		return "Composing final video"
	case "complete":
		return "Complete"
	case "failed":
		return "Failed"
	default:
		// Handle scene stages (scene_N_generating, scene_N_complete)
		if len(stage) > 6 && stage[:6] == "scene_" {
			var sceneNum int
			var suffix string
			_, err := fmt.Sscanf(stage, "scene_%d_%s", &sceneNum, &suffix)
			if err == nil {
				switch suffix {
				case "generating":
					return fmt.Sprintf("Generating scene %d", sceneNum)
				case "complete":
					return fmt.Sprintf("Scene %d ready", sceneNum)
				}
			}
		}
		return stage
	}
}

// buildStagesCompleted builds list of completed stages based on current stage
func buildStagesCompleted(job *domain.Job) []StageInfo {
	stages := make([]StageInfo, 0)
	currentStage := job.Stage

	// Script generation
	if currentStage != "script_generating" {
		stages = append(stages, StageInfo{
			Name:        "script_complete",
			DisplayName: "Script ready",
			Progress:    15,
			CompletedAt: nil, // Could track completion times if needed
		})
	}

	// Scenes
	for i := 0; i < job.ScenesCompleted; i++ {
		stageName := fmt.Sprintf("scene_%d_complete", i+1)
		progress := calculateDynamicProgress(stageName, len(job.Scenes))
		stages = append(stages, StageInfo{
			Name:        stageName,
			DisplayName: fmt.Sprintf("Scene %d ready", i+1),
			Progress:    progress,
		})
	}

	// Audio (only if stage is past audio)
	if currentStage == "audio_complete" || currentStage == "composing" || currentStage == "complete" {
		stages = append(stages, StageInfo{
			Name:        "audio_complete",
			DisplayName: "Audio ready",
			Progress:    95,
		})
	}

	// Composition (only if complete)
	if currentStage == "complete" {
		stages = append(stages, StageInfo{
			Name:        "complete",
			DisplayName: "Complete",
			Progress:    100,
			CompletedAt: job.CompletedAt,
		})
	}

	return stages
}

// buildStagesPending builds list of pending stages based on current stage
func buildStagesPending(job *domain.Job) []StageInfo {
	stages := make([]StageInfo, 0)
	currentStage := job.Stage
	totalScenes := len(job.Scenes)

	// Scenes
	for i := job.ScenesCompleted; i < totalScenes; i++ {
		stageName := fmt.Sprintf("scene_%d_generating", i+1)
		progress := calculateDynamicProgress(stageName, totalScenes)
		stages = append(stages, StageInfo{
			Name:        stageName,
			DisplayName: fmt.Sprintf("Generating scene %d", i+1),
			Progress:    progress,
		})
	}

	// Audio (if not yet started)
	if currentStage != "audio_generating" && currentStage != "audio_complete" &&
		currentStage != "composing" && currentStage != "complete" {
		stages = append(stages, StageInfo{
			Name:        "audio_generating",
			DisplayName: "Generating background music",
			Progress:    85,
		})
	}

	// Composition (if not yet started)
	if currentStage != "composing" && currentStage != "complete" {
		stages = append(stages, StageInfo{
			Name:        "composing",
			DisplayName: "Composing final video",
			Progress:    95,
		})
	}

	return stages
}

// calculateETA estimates time remaining based on elapsed time and current progress
func calculateETA(stage string, startTime time.Time, totalScenes int) int {
	progress := calculateDynamicProgress(stage, totalScenes)

	if progress == 0 || progress >= 100 {
		return 0
	}

	elapsed := time.Since(startTime).Seconds()
	totalEstimated := elapsed * 100.0 / float64(progress)
	remaining := totalEstimated - elapsed

	if remaining < 0 {
		return 0
	}

	return int(remaining)
}
