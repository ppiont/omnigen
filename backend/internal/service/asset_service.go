package service

import (
	"context"
	"time"

	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"go.uber.org/zap"
)

// AssetService handles asset URL generation and management
type AssetService struct {
	s3Repo *repository.S3AssetRepository
	logger *zap.Logger
}

// NewAssetService creates a new asset service instance
func NewAssetService(
	s3Repo *repository.S3AssetRepository,
	logger *zap.Logger,
) *AssetService {
	return &AssetService{
		s3Repo: s3Repo,
		logger: logger,
	}
}

// JobAssets contains all presigned URLs for a job's assets
type JobAssets struct {
	SceneClips    []string `json:"scene_clips"`
	Thumbnails    []string `json:"thumbnails"`
	AudioURL      string   `json:"audio_url,omitempty"`
	FinalVideoURL string   `json:"final_video_url,omitempty"`
}

// GetJobAssets generates presigned URLs for all assets associated with a job
func (s *AssetService) GetJobAssets(ctx context.Context, job *domain.Job, duration time.Duration) (*JobAssets, error) {
	assets := &JobAssets{
		SceneClips: make([]string, 0, len(job.SceneVideoURLs)),
		Thumbnails: make([]string, 0, len(job.Scenes)),
	}

	s.logger.Info("Generating presigned URLs for job assets",
		zap.String("job_id", job.JobID),
		zap.Int("scene_count", len(job.SceneVideoURLs)),
		zap.Duration("url_duration", duration),
	)

	// Generate presigned URLs for scene clips
	for i, sceneURL := range job.SceneVideoURLs {
		key := extractS3Key(sceneURL)
		url, err := s.s3Repo.GetPresignedURL(ctx, key, duration)
		if err != nil {
			s.logger.Warn("Failed to generate presigned URL for scene clip",
				zap.String("job_id", job.JobID),
				zap.Int("scene_number", i+1),
				zap.String("key", key),
				zap.Error(err),
			)
			continue // Skip this clip but continue processing others
		}
		assets.SceneClips = append(assets.SceneClips, url)
	}

	// Generate presigned URL for thumbnail (last scene thumbnail)
	if job.ThumbnailURL != "" {
		key := extractS3Key(job.ThumbnailURL)
		url, err := s.s3Repo.GetPresignedURL(ctx, key, duration)
		if err != nil {
			s.logger.Warn("Failed to generate presigned URL for thumbnail",
				zap.String("job_id", job.JobID),
				zap.String("key", key),
				zap.Error(err),
			)
		} else {
			// For now, put single thumbnail in array (could expand to all scene thumbnails later)
			assets.Thumbnails = append(assets.Thumbnails, url)
		}
	}

	// Generate presigned URL for audio
	if job.AudioURL != "" {
		key := extractS3Key(job.AudioURL)
		url, err := s.s3Repo.GetPresignedURL(ctx, key, duration)
		if err != nil {
			s.logger.Warn("Failed to generate presigned URL for audio",
				zap.String("job_id", job.JobID),
				zap.String("key", key),
				zap.Error(err),
			)
		} else {
			assets.AudioURL = url
		}
	}

	// Generate presigned URL for final video
	if job.Status == domain.StatusCompleted && job.VideoKey != "" {
		url, err := s.s3Repo.GetPresignedURL(ctx, job.VideoKey, duration)
		if err != nil {
			s.logger.Warn("Failed to generate presigned URL for final video",
				zap.String("job_id", job.JobID),
				zap.String("key", job.VideoKey),
				zap.Error(err),
			)
		} else {
			assets.FinalVideoURL = url
		}
	}

	s.logger.Info("Presigned URLs generated successfully",
		zap.String("job_id", job.JobID),
		zap.Int("scene_clips", len(assets.SceneClips)),
		zap.Int("thumbnails", len(assets.Thumbnails)),
		zap.Bool("has_audio", assets.AudioURL != ""),
		zap.Bool("has_final_video", assets.FinalVideoURL != ""),
	)

	return assets, nil
}

// extractS3Key extracts the S3 key from a full S3 URL
// Input: https://bucket.s3.amazonaws.com/users/user123/jobs/job456/clips/scene-001.mp4
// Output: users/user123/jobs/job456/clips/scene-001.mp4
func extractS3Key(s3URL string) string {
	// This function already exists in generate_async.go (line 636)
	// We're duplicating it here for the service layer
	// Could be refactored to a shared utility package later

	// Simple implementation: find ".amazonaws.com/" and return everything after it
	const marker = ".amazonaws.com/"
	idx := findSubstring(s3URL, marker)
	if idx == -1 {
		// Fallback: assume it's already just a key
		return s3URL
	}
	return s3URL[idx+len(marker):]
}

// findSubstring returns the index of substr in s, or -1 if not found
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
