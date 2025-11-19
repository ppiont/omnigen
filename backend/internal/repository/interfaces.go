package repository

import (
	"context"
	"time"

	"github.com/omnigen/backend/internal/domain"
)

// JobRepository defines the interface for job persistence operations
type JobRepository interface {
	// CreateJob creates a new job record
	CreateJob(ctx context.Context, job *domain.Job) error

	// GetJob retrieves a job by ID
	GetJob(ctx context.Context, jobID string) (*domain.Job, error)

	// GetJobsByUser retrieves all jobs for a user, optionally filtered by status
	GetJobsByUser(ctx context.Context, userID string, limit int, status string) ([]*domain.Job, error)

	// UpdateJobStageWithMetadata updates stage and metadata atomically
	UpdateJobStageWithMetadata(ctx context.Context, jobID string, stage string, metadata map[string]interface{}) error

	// MarkJobComplete marks a job as completed with video key
	MarkJobComplete(ctx context.Context, jobID string, videoKey string) error

	// MarkJobFailed marks a job as failed with error message
	MarkJobFailed(ctx context.Context, jobID string, errorMsg string) error

	// UpdateJob updates an entire job record atomically
	UpdateJob(ctx context.Context, job *domain.Job) error

	// HealthCheck verifies the repository is operational
	HealthCheck(ctx context.Context) error
}

// AssetRepository defines the interface for asset storage operations
type AssetRepository interface {
	// GetPresignedURL generates a presigned URL for downloading an asset
	GetPresignedURL(ctx context.Context, key string, duration time.Duration) (string, error)

	// UploadFile uploads a file to storage
	UploadFile(ctx context.Context, bucket, key, filePath string, contentType string) (string, error)

	// DownloadFile downloads a file from storage
	DownloadFile(ctx context.Context, bucket, key, destPath string) error

	// HealthCheck verifies the repository is operational
	HealthCheck(ctx context.Context) error
}

// UsageRepository defines the interface for usage tracking operations
type UsageRepository interface {
	// GetOrCreateUsage retrieves or creates a usage record for a user
	GetOrCreateUsage(ctx context.Context, userID, subscriptionTier string) (*domain.Usage, error)

	// CheckAndDecrementQuota checks quota and decrements if available
	CheckAndDecrementQuota(ctx context.Context, userID, subscriptionTier string) error

	// IncrementUsage increments usage counters
	IncrementUsage(ctx context.Context, userID string, videoDuration int) error
}
