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
	// ErrBrandGuidelinesNotFound is returned when brand guidelines are not found
	ErrBrandGuidelinesNotFound = errors.New("brand guidelines not found")
)

// DynamoDBBrandGuidelinesRepository handles DynamoDB operations for brand guidelines
type DynamoDBBrandGuidelinesRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    *zap.Logger
}

// NewDynamoDBBrandGuidelinesRepository creates a new DynamoDB brand guidelines repository
func NewDynamoDBBrandGuidelinesRepository(
	client *dynamodb.Client,
	tableName string,
	logger *zap.Logger,
) *DynamoDBBrandGuidelinesRepository {
	return &DynamoDBBrandGuidelinesRepository{
		client:    client,
		tableName: tableName,
		logger:    logger,
	}
}

// CreateBrandGuidelines creates new brand guidelines in DynamoDB
func (r *DynamoDBBrandGuidelinesRepository) CreateBrandGuidelines(ctx context.Context, guidelines *domain.BrandGuidelines) error {
	item, err := attributevalue.MarshalMap(guidelines)
	if err != nil {
		r.logger.Error("Failed to marshal brand guidelines", zap.Error(err))
		return fmt.Errorf("failed to marshal brand guidelines: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
		// Ensure we don't overwrite existing guidelines with same ID
		ConditionExpression: aws.String("attribute_not_exists(guideline_id)"),
	})

	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return fmt.Errorf("brand guidelines with ID %s already exists", guidelines.GuidelineID)
		}
		r.logger.Error("Failed to create brand guidelines",
			zap.String("guideline_id", guidelines.GuidelineID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create brand guidelines: %w", err)
	}

	return nil
}

// GetBrandGuidelines retrieves brand guidelines by ID
func (r *DynamoDBBrandGuidelinesRepository) GetBrandGuidelines(ctx context.Context, guidelineID string) (*domain.BrandGuidelines, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"guideline_id": &types.AttributeValueMemberS{Value: guidelineID},
		},
	})

	if err != nil {
		r.logger.Error("Failed to get brand guidelines",
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get brand guidelines: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var guidelines domain.BrandGuidelines
	if err := attributevalue.UnmarshalMap(result.Item, &guidelines); err != nil {
		r.logger.Error("Failed to unmarshal brand guidelines",
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to unmarshal brand guidelines: %w", err)
	}

	return &guidelines, nil
}

// GetBrandGuidelinesByUser retrieves all brand guidelines for a specific user
func (r *DynamoDBBrandGuidelinesRepository) GetBrandGuidelinesByUser(ctx context.Context, userID string) ([]*domain.BrandGuidelines, error) {
	// Since we need to query by user_id which is not the primary key, we need a GSI
	// For now, we'll scan the table with a filter (not ideal for production, but works for MVP)
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id": &types.AttributeValueMemberS{Value: userID},
		},
	})

	if err != nil {
		r.logger.Error("Failed to scan brand guidelines by user",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get brand guidelines by user: %w", err)
	}

	guidelines := make([]*domain.BrandGuidelines, 0, len(result.Items))
	for _, item := range result.Items {
		var guideline domain.BrandGuidelines
		if err := attributevalue.UnmarshalMap(item, &guideline); err != nil {
			r.logger.Error("Failed to unmarshal brand guideline item",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			continue // Skip invalid items
		}
		guidelines = append(guidelines, &guideline)
	}

	return guidelines, nil
}

// GetActiveBrandGuidelines retrieves the active brand guidelines for a user
func (r *DynamoDBBrandGuidelinesRepository) GetActiveBrandGuidelines(ctx context.Context, userID string) (*domain.BrandGuidelines, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("user_id = :user_id AND is_active = :is_active"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id":    &types.AttributeValueMemberS{Value: userID},
			":is_active": &types.AttributeValueMemberBOOL{Value: true},
		},
		Limit: aws.Int32(1), // Only expect one active guideline per user
	})

	if err != nil {
		r.logger.Error("Failed to get active brand guidelines",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get active brand guidelines: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, nil // No active guidelines found
	}

	var guidelines domain.BrandGuidelines
	if err := attributevalue.UnmarshalMap(result.Items[0], &guidelines); err != nil {
		r.logger.Error("Failed to unmarshal active brand guidelines",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to unmarshal active brand guidelines: %w", err)
	}

	return &guidelines, nil
}

// UpdateBrandGuidelines updates existing brand guidelines
func (r *DynamoDBBrandGuidelinesRepository) UpdateBrandGuidelines(ctx context.Context, guidelines *domain.BrandGuidelines) error {
	item, err := attributevalue.MarshalMap(guidelines)
	if err != nil {
		r.logger.Error("Failed to marshal brand guidelines for update", zap.Error(err))
		return fmt.Errorf("failed to marshal brand guidelines: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
		// Ensure the item exists before updating
		ConditionExpression: aws.String("attribute_exists(guideline_id)"),
	})

	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return ErrBrandGuidelinesNotFound
		}
		r.logger.Error("Failed to update brand guidelines",
			zap.String("guideline_id", guidelines.GuidelineID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update brand guidelines: %w", err)
	}

	return nil
}

// SetActiveBrandGuidelines sets a specific guideline as active and deactivates others for the user
func (r *DynamoDBBrandGuidelinesRepository) SetActiveBrandGuidelines(ctx context.Context, userID, guidelineID string) error {
	// First, get all guidelines for the user
	allGuidelines, err := r.GetBrandGuidelinesByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user guidelines: %w", err)
	}

	// Find the target guideline and verify it exists
	var targetGuideline *domain.BrandGuidelines
	for _, g := range allGuidelines {
		if g.GuidelineID == guidelineID {
			targetGuideline = g
			break
		}
	}

	if targetGuideline == nil {
		return ErrBrandGuidelinesNotFound
	}

	// Update all guidelines for this user
	for _, g := range allGuidelines {
		originalActive := g.IsActive
		g.IsActive = (g.GuidelineID == guidelineID)

		// Only update if the active status changed
		if originalActive != g.IsActive {
			if err := r.UpdateBrandGuidelines(ctx, g); err != nil {
				r.logger.Error("Failed to update guideline active status",
					zap.String("guideline_id", g.GuidelineID),
					zap.Bool("new_active", g.IsActive),
					zap.Error(err),
				)
				return fmt.Errorf("failed to update guideline active status: %w", err)
			}
		}
	}

	return nil
}

// DeleteBrandGuidelines deletes brand guidelines by ID
func (r *DynamoDBBrandGuidelinesRepository) DeleteBrandGuidelines(ctx context.Context, guidelineID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"guideline_id": &types.AttributeValueMemberS{Value: guidelineID},
		},
		// Ensure the item exists before deleting
		ConditionExpression: aws.String("attribute_exists(guideline_id)"),
	})

	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return ErrBrandGuidelinesNotFound
		}
		r.logger.Error("Failed to delete brand guidelines",
			zap.String("guideline_id", guidelineID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete brand guidelines: %w", err)
	}

	return nil
}

// HealthCheck verifies the repository is operational
func (r *DynamoDBBrandGuidelinesRepository) HealthCheck(ctx context.Context) error {
	_, err := r.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(r.tableName),
	})

	if err != nil {
		r.logger.Error("Brand guidelines repository health check failed", zap.Error(err))
		return fmt.Errorf("brand guidelines repository health check failed: %w", err)
	}

	return nil
}