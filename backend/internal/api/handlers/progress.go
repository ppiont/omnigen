package handlers

import (
	"fmt"
	"net/http"
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

	h.logger.Info("Getting progress for job",
		zap.String("job_id", jobID),
	)

	// Fetch job from DynamoDB
	job, err := h.jobRepo.GetJob(c.Request.Context(), jobID)
	if err != nil {
		if err.Error() == "job not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Job not found",
			})
			return
		}
		h.logger.Error("Failed to fetch job",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch job",
		})
		return
	}

	// Calculate progress percentage and ETA
	progress := calculateDynamicProgress(job.Stage, len(job.Scenes))
	eta := calculateETA(job.Stage, time.Unix(job.CreatedAt, 0), len(job.Scenes))

	// Generate presigned URLs for all assets
	assets, err := h.assetService.GetJobAssets(c.Request.Context(), job, 1*time.Hour)
	if err != nil {
		h.logger.Warn("Failed to generate asset URLs",
			zap.String("job_id", jobID),
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
	response := ProgressResponse{
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

	h.logger.Info("Progress retrieved successfully",
		zap.String("job_id", jobID),
		zap.Int("progress", progress),
		zap.Int("eta_seconds", eta),
		zap.String("status", job.Status),
	)

	c.JSON(http.StatusOK, response)
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
