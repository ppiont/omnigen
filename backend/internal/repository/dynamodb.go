package repository

import (
	"context"
	"errors"
	"fmt"

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
