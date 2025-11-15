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
	// ErrUsageNotFound is returned when usage record is not found
	ErrUsageNotFound = errors.New("usage not found")
	// ErrQuotaExceeded is returned when user has no remaining quota
	ErrQuotaExceeded = errors.New("quota exceeded")
)

// Subscription tier quotas (videos per month)
var tierQuotas = map[string]int{
	"free":       10,
	"pro":        100,
	"enterprise": 1000,
}

// UsageRepository handles DynamoDB operations for usage tracking
type UsageRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    *zap.Logger
}

// NewUsageRepository creates a new usage repository
func NewUsageRepository(
	client *dynamodb.Client,
	tableName string,
	logger *zap.Logger,
) *UsageRepository {
	return &UsageRepository{
		client:    client,
		tableName: tableName,
		logger:    logger,
	}
}

// GetCurrentPeriod returns the current billing period (YYYY-MM format)
func GetCurrentPeriod() string {
	now := time.Now()
	return fmt.Sprintf("%d-%02d", now.Year(), now.Month())
}

// GetOrCreateUsage retrieves or creates usage record for user and current period
func (r *UsageRepository) GetOrCreateUsage(ctx context.Context, userID, subscriptionTier string) (*domain.Usage, error) {
	period := GetCurrentPeriod()

	// Try to get existing usage record
	usage, err := r.GetUsage(ctx, userID, period)
	if err == nil {
		return usage, nil
	}

	if !errors.Is(err, ErrUsageNotFound) {
		return nil, err
	}

	// Create new usage record for the period
	quota := tierQuotas[subscriptionTier]
	if quota == 0 {
		quota = tierQuotas["free"] // Default to free tier
	}

	usage = &domain.Usage{
		UserID:         userID,
		Period:         period,
		RequestCount:   0,
		VideoGenerated: 0,
		TotalDuration:  0,
		MonthlyQuota:   quota,
		QuotaRemaining: quota,
		LastUpdated:    time.Now(),
	}

	if err := r.CreateUsage(ctx, usage); err != nil {
		return nil, err
	}

	return usage, nil
}

// GetUsage retrieves usage record for a specific user and period
func (r *UsageRepository) GetUsage(ctx context.Context, userID, period string) (*domain.Usage, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
			"period":  &types.AttributeValueMemberS{Value: period},
		},
	})
	if err != nil {
		r.logger.Error("Failed to get usage",
			zap.String("user_id", userID),
			zap.String("period", period),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	if result.Item == nil {
		return nil, ErrUsageNotFound
	}

	var usage domain.Usage
	if err := attributevalue.UnmarshalMap(result.Item, &usage); err != nil {
		r.logger.Error("Failed to unmarshal usage",
			zap.String("user_id", userID),
			zap.String("period", period),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to unmarshal usage: %w", err)
	}

	return &usage, nil
}

// CreateUsage creates a new usage record
func (r *UsageRepository) CreateUsage(ctx context.Context, usage *domain.Usage) error {
	item, err := attributevalue.MarshalMap(usage)
	if err != nil {
		r.logger.Error("Failed to marshal usage", zap.Error(err))
		return fmt.Errorf("failed to marshal usage: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		r.logger.Error("Failed to create usage",
			zap.String("user_id", usage.UserID),
			zap.String("period", usage.Period),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create usage: %w", err)
	}

	r.logger.Info("Usage record created",
		zap.String("user_id", usage.UserID),
		zap.String("period", usage.Period),
	)
	return nil
}

// CheckAndDecrementQuota checks if user has remaining quota and decrements it
// Returns error if quota is exceeded
func (r *UsageRepository) CheckAndDecrementQuota(ctx context.Context, userID, subscriptionTier string) error {
	period := GetCurrentPeriod()

	// Get or create usage record
	usage, err := r.GetOrCreateUsage(ctx, userID, subscriptionTier)
	if err != nil {
		return err
	}

	// Check if user has remaining quota
	if !usage.HasQuotaRemaining() {
		r.logger.Warn("User quota exceeded",
			zap.String("user_id", userID),
			zap.String("period", period),
			zap.Int("quota_remaining", usage.QuotaRemaining),
		)
		return ErrQuotaExceeded
	}

	// Decrement quota atomically
	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
			"period":  &types.AttributeValueMemberS{Value: period},
		},
		UpdateExpression: aws.String("SET quota_remaining = quota_remaining - :dec, last_updated = :now"),
		ConditionExpression: aws.String("quota_remaining > :zero"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":dec":  &types.AttributeValueMemberN{Value: "1"},
			":zero": &types.AttributeValueMemberN{Value: "0"},
			":now":  &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	if err != nil {
		// Check if it's a condition check failure (quota exhausted)
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			r.logger.Warn("Quota exhausted during decrement",
				zap.String("user_id", userID),
				zap.String("period", period),
			)
			return ErrQuotaExceeded
		}

		r.logger.Error("Failed to decrement quota",
			zap.String("user_id", userID),
			zap.String("period", period),
			zap.Error(err),
		)
		return fmt.Errorf("failed to decrement quota: %w", err)
	}

	r.logger.Info("Quota decremented",
		zap.String("user_id", userID),
		zap.String("period", period),
	)
	return nil
}

// IncrementUsage increments usage counters after successful video generation
func (r *UsageRepository) IncrementUsage(ctx context.Context, userID string, videoDuration int) error {
	period := GetCurrentPeriod()

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
			"period":  &types.AttributeValueMemberS{Value: period},
		},
		UpdateExpression: aws.String("ADD request_count :one, video_generated :one, total_duration :duration SET last_updated = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":one":      &types.AttributeValueMemberN{Value: "1"},
			":duration": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", videoDuration)},
			":now":      &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	if err != nil {
		r.logger.Error("Failed to increment usage",
			zap.String("user_id", userID),
			zap.String("period", period),
			zap.Int("duration", videoDuration),
			zap.Error(err),
		)
		return fmt.Errorf("failed to increment usage: %w", err)
	}

	r.logger.Info("Usage incremented",
		zap.String("user_id", userID),
		zap.String("period", period),
		zap.Int("duration", videoDuration),
	)
	return nil
}
