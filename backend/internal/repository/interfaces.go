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

	// MarkJobComplete marks a job as completed with video keys (MP4 required, WebM optional)
	MarkJobComplete(ctx context.Context, jobID string, videoKey string, webmVideoKey ...string) error

	// MarkJobFailed marks a job as failed with error message
	MarkJobFailed(ctx context.Context, jobID string, errorMsg string) error

	// UpdateJob updates an entire job record atomically
	UpdateJob(ctx context.Context, job *domain.Job) error

	// DeleteJob deletes a job by ID
	DeleteJob(ctx context.Context, jobID string) error

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

	// DeletePrefix deletes all assets under a prefix (best-effort cleanup)
	DeletePrefix(ctx context.Context, bucket, prefix string) error

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

// BrandGuidelinesRepository defines the interface for brand guidelines persistence operations
type BrandGuidelinesRepository interface {
	// CreateBrandGuidelines creates a new brand guidelines record
	CreateBrandGuidelines(ctx context.Context, guidelines *domain.BrandGuidelines) error

	// GetBrandGuidelines retrieves brand guidelines by ID
	GetBrandGuidelines(ctx context.Context, guidelineID string) (*domain.BrandGuidelines, error)

	// GetBrandGuidelinesByUser retrieves all brand guidelines for a user
	GetBrandGuidelinesByUser(ctx context.Context, userID string) ([]*domain.BrandGuidelines, error)

	// GetActiveBrandGuidelines retrieves the active brand guidelines for a user
	GetActiveBrandGuidelines(ctx context.Context, userID string) (*domain.BrandGuidelines, error)

	// UpdateBrandGuidelines updates brand guidelines
	UpdateBrandGuidelines(ctx context.Context, guidelines *domain.BrandGuidelines) error

	// SetActiveBrandGuidelines sets a specific guideline as active and deactivates others for the user
	SetActiveBrandGuidelines(ctx context.Context, userID, guidelineID string) error

	// DeleteBrandGuidelines deletes brand guidelines by ID
	DeleteBrandGuidelines(ctx context.Context, guidelineID string) error

	// HealthCheck verifies the repository is operational
	HealthCheck(ctx context.Context) error
}
