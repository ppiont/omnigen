package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/omnigen/backend/internal/domain"
	"go.uber.org/zap"
)

var (
	// ErrJobNotFound is returned when a job is not found
	ErrJobNotFound = errors.New("job not found")
)

// DynamoDBRepository handles DynamoDB operations for jobs
type DynamoDBRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    *zap.Logger
}

// NewDynamoDBRepository creates a new DynamoDB repository
func NewDynamoDBRepository(
	client *dynamodb.Client,
	tableName string,
	logger *zap.Logger,
) *DynamoDBRepository {
	return &DynamoDBRepository{
		client:    client,
		tableName: tableName,
		logger:    logger,
	}
}

// CreateJob creates a new job in DynamoDB
func (r *DynamoDBRepository) CreateJob(ctx context.Context, job *domain.Job) error {
	item, err := attributevalue.MarshalMap(job)
	if err != nil {
		r.logger.Error("Failed to marshal job", zap.Error(err))
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		r.logger.Error("Failed to put item", zap.String("job_id", job.JobID), zap.Error(err))
		return fmt.Errorf("failed to put item: %w", err)
	}

	r.logger.Info("Job created successfully", zap.String("job_id", job.JobID))
	return nil
}

// GetJob retrieves a job by ID
func (r *DynamoDBRepository) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"job_id": &types.AttributeValueMemberS{Value: jobID},
		},
	})
	if err != nil {
		r.logger.Error("Failed to get item", zap.String("job_id", jobID), zap.Error(err))
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if result.Item == nil {
		return nil, ErrJobNotFound
	}

	var job domain.Job
	if err := attributevalue.UnmarshalMap(result.Item, &job); err != nil {
		r.logger.Error("Failed to unmarshal job", zap.String("job_id", jobID), zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// UpdateJobStatus updates the status of a job
func (r *DynamoDBRepository) UpdateJobStatus(ctx context.Context, jobID string, status string) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"job_id": &types.AttributeValueMemberS{Value: jobID},
		},
		UpdateExpression: aws.String("SET #status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: status},
		},
	})
	if err != nil {
		r.logger.Error("Failed to update job status",
			zap.String("job_id", jobID),
			zap.String("status", status),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// UpdateJob updates an entire job record
func (r *DynamoDBRepository) UpdateJob(ctx context.Context, job *domain.Job) error {
	item, err := attributevalue.MarshalMap(job)
	if err != nil {
		r.logger.Error("Failed to marshal job", zap.Error(err))
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		r.logger.Error("Failed to update job", zap.String("job_id", job.JobID), zap.Error(err))
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

// UpdateJobStage updates the stage and updated_at timestamp
func (r *DynamoDBRepository) UpdateJobStage(ctx context.Context, jobID string, stage string) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"job_id": &types.AttributeValueMemberS{Value: jobID},
		},
		UpdateExpression: aws.String("SET #stage = :stage, #updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#stage":      "stage",
			"#updated_at": "updated_at",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":stage":      &types.AttributeValueMemberS{Value: stage},
			":updated_at": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", getCurrentTimestamp())},
		},
	})
	if err != nil {
		r.logger.Error("Failed to update job stage",
			zap.String("job_id", jobID),
			zap.String("stage", stage),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update job stage: %w", err)
	}

	return nil
}

// UpdateJobStageWithMetadata atomically updates stage and metadata
func (r *DynamoDBRepository) UpdateJobStageWithMetadata(
	ctx context.Context,
	jobID string,
	stage string,
	metadata map[string]interface{},
) error {
	metadataAttr, err := attributevalue.Marshal(metadata)
	if err != nil {
		r.logger.Error("Failed to marshal metadata", zap.Error(err))
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"job_id": &types.AttributeValueMemberS{Value: jobID},
		},
		UpdateExpression: aws.String("SET #stage = :stage, #metadata = :metadata, #updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#stage":      "stage",
			"#metadata":   "metadata",
			"#updated_at": "updated_at",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":stage":      &types.AttributeValueMemberS{Value: stage},
			":metadata":   metadataAttr,
			":updated_at": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", getCurrentTimestamp())},
		},
	})
	if err != nil {
		r.logger.Error("Failed to update job stage with metadata",
			zap.String("job_id", jobID),
			zap.String("stage", stage),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update job stage with metadata: %w", err)
	}

	return nil
}

// MarkJobComplete marks a job as completed with video URL
func (r *DynamoDBRepository) MarkJobComplete(ctx context.Context, jobID string, videoKey string) error {
	now := getCurrentTimestamp()

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"job_id": &types.AttributeValueMemberS{Value: jobID},
		},
		UpdateExpression: aws.String("SET #status = :status, #stage = :stage, #video_key = :video_key, #completed_at = :completed_at, #updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#status":       "status",
			"#stage":        "stage",
			"#video_key":    "video_key",
			"#completed_at": "completed_at",
			"#updated_at":   "updated_at",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":       &types.AttributeValueMemberS{Value: domain.StatusCompleted},
			":stage":        &types.AttributeValueMemberS{Value: "complete"},
			":video_key":    &types.AttributeValueMemberS{Value: videoKey},
			":completed_at": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
			":updated_at":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
		},
	})
	if err != nil {
		r.logger.Error("Failed to mark job complete",
			zap.String("job_id", jobID),
			zap.String("video_key", videoKey),
			zap.Error(err),
		)
		return fmt.Errorf("failed to mark job complete: %w", err)
	}

	r.logger.Info("Job marked as complete",
		zap.String("job_id", jobID),
		zap.String("video_key", videoKey),
	)
	return nil
}

// MarkJobFailed marks a job as failed with error message
func (r *DynamoDBRepository) MarkJobFailed(ctx context.Context, jobID string, errorMsg string) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"job_id": &types.AttributeValueMemberS{Value: jobID},
		},
		UpdateExpression: aws.String("SET #status = :status, #stage = :stage, #error_message = :error_message, #updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#status":        "status",
			"#stage":         "stage",
			"#error_message": "error_message",
			"#updated_at":    "updated_at",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":        &types.AttributeValueMemberS{Value: domain.StatusFailed},
			":stage":         &types.AttributeValueMemberS{Value: "failed"},
			":error_message": &types.AttributeValueMemberS{Value: errorMsg},
			":updated_at":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", getCurrentTimestamp())},
		},
	})
	if err != nil {
		r.logger.Error("Failed to mark job as failed",
			zap.String("job_id", jobID),
			zap.String("error", errorMsg),
			zap.Error(err),
		)
		return fmt.Errorf("failed to mark job failed: %w", err)
	}

	r.logger.Warn("Job marked as failed",
		zap.String("job_id", jobID),
		zap.String("error", errorMsg),
	)
	return nil
}

// ListJobsByUser retrieves jobs for a specific user with pagination
func (r *DynamoDBRepository) ListJobsByUser(ctx context.Context, userID string, limit int, lastEvaluatedKey map[string]types.AttributeValue) ([]domain.Job, map[string]types.AttributeValue, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("user-jobs-index"),
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id": &types.AttributeValueMemberS{Value: userID},
		},
		ScanIndexForward: aws.Bool(false), // Sort descending (newest first)
		Limit:            aws.Int32(int32(limit)),
	}

	if lastEvaluatedKey != nil {
		input.ExclusiveStartKey = lastEvaluatedKey
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		r.logger.Error("Failed to query jobs",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, nil, fmt.Errorf("failed to query jobs: %w", err)
	}

	var jobs []domain.Job
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &jobs); err != nil {
		r.logger.Error("Failed to unmarshal jobs",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, nil, fmt.Errorf("failed to unmarshal jobs: %w", err)
	}

	return jobs, result.LastEvaluatedKey, nil
}

// DeleteJob deletes a job from DynamoDB
func (r *DynamoDBRepository) DeleteJob(ctx context.Context, jobID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"job_id": &types.AttributeValueMemberS{Value: jobID},
		},
	})
	if err != nil {
		r.logger.Error("Failed to delete job",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete job: %w", err)
	}

	r.logger.Info("Job deleted successfully", zap.String("job_id", jobID))
	return nil
}

// getCurrentTimestamp returns the current Unix timestamp
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// HealthCheck performs a lightweight health check on DynamoDB
func (r *DynamoDBRepository) HealthCheck(ctx context.Context) error {
	_, err := r.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(r.tableName),
	})
	if err != nil {
		return fmt.Errorf("dynamodb health check failed: %w", err)
	}
	return nil
}
