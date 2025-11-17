package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/omnigen/backend/internal/adapters"
	"github.com/omnigen/backend/internal/domain"
)

// ParserService handles ad script generation and management
type ParserService struct {
	gpt4o        *adapters.GPT4oAdapter
	dynamoDB     *dynamodb.Client
	scriptsTable string
	logger       *zap.Logger
}

// ParseRequest represents user input for script generation - SIMPLE interface
type ParseRequest struct {
	UserID      string `json:"user_id"`
	Prompt      string `json:"prompt"`                // Free-form user input with ALL context
	Duration    int    `json:"duration"`              // 10-60 seconds (must be multiple of 10)
	AspectRatio string `json:"aspect_ratio"`          // "16:9", "9:16", or "1:1"
	StartImage  string `json:"start_image,omitempty"` // Optional starting image URL
}

// NewParserService creates a new script parser service
func NewParserService(
	gpt4o *adapters.GPT4oAdapter,
	dynamoDB *dynamodb.Client,
	scriptsTable string,
	logger *zap.Logger,
) *ParserService {
	return &ParserService{
		gpt4o:        gpt4o,
		dynamoDB:     dynamoDB,
		scriptsTable: scriptsTable,
		logger:       logger,
	}
}

// GenerateScript creates a new ad script using GPT-4o
func (s *ParserService) GenerateScript(ctx context.Context, req ParseRequest) (*domain.Script, error) {
	s.logger.Info("Generating script with GPT-4o",
		zap.String("user_id", req.UserID),
		zap.Int("duration", req.Duration),
		zap.String("prompt", req.Prompt))

	// Validate request
	if err := s.validateParseRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Call GPT-4o adapter - GPT-4o will extract product info from prompt
	gpt4oReq := &adapters.ScriptGenerationRequest{
		Prompt:      req.Prompt,
		Duration:    req.Duration,
		AspectRatio: req.AspectRatio,
		StartImage:  req.StartImage,
	}

	script, err := s.gpt4o.GenerateScript(ctx, gpt4oReq)
	if err != nil {
		return nil, fmt.Errorf("GPT-4o generation failed: %w", err)
	}

	// Populate metadata
	script.ScriptID = fmt.Sprintf("script-%s", uuid.New().String())
	script.UserID = req.UserID
	script.CreatedAt = time.Now().Unix()
	script.UpdatedAt = time.Now().Unix()
	script.Status = "generated"
	script.ExpiresAt = time.Now().Add(30 * 24 * time.Hour).Unix() // 30-day TTL

	// Save to DynamoDB
	if err := s.SaveScript(ctx, script); err != nil {
		return nil, fmt.Errorf("failed to save script: %w", err)
	}

	s.logger.Info("Script generated successfully",
		zap.String("script_id", script.ScriptID),
		zap.Int("num_scenes", len(script.Scenes)))

	return script, nil
}

// GetScript retrieves a script by ID
func (s *ParserService) GetScript(ctx context.Context, scriptID string) (*domain.Script, error) {
	result, err := s.dynamoDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.scriptsTable),
		Key: map[string]types.AttributeValue{
			"script_id": &types.AttributeValueMemberS{Value: scriptID},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get script: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("script not found: %s", scriptID)
	}

	var script domain.Script
	if err := attributevalue.UnmarshalMap(result.Item, &script); err != nil {
		return nil, fmt.Errorf("failed to unmarshal script: %w", err)
	}

	return &script, nil
}

// UpdateScript updates an existing script
func (s *ParserService) UpdateScript(ctx context.Context, script *domain.Script) error {
	// Update timestamp
	script.UpdatedAt = time.Now().Unix()

	// Validate before saving
	if err := s.validateScript(script); err != nil {
		return fmt.Errorf("script validation failed: %w", err)
	}

	return s.SaveScript(ctx, script)
}

// SaveScript saves a script to DynamoDB
func (s *ParserService) SaveScript(ctx context.Context, script *domain.Script) error {
	item, err := attributevalue.MarshalMap(script)
	if err != nil {
		return fmt.Errorf("failed to marshal script: %w", err)
	}

	_, err = s.dynamoDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.scriptsTable),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to save script to DynamoDB: %w", err)
	}

	s.logger.Info("Script saved", zap.String("script_id", script.ScriptID))
	return nil
}

// validateParseRequest validates the parse request
func (s *ParserService) validateParseRequest(req ParseRequest) error {
	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	if req.Duration < 10 || req.Duration > 60 {
		return fmt.Errorf("duration must be between 10 and 60 seconds")
	}

	if req.Duration%10 != 0 {
		return fmt.Errorf("duration must be a multiple of 10 seconds (Kling constraint)")
	}

	if req.AspectRatio != "16:9" && req.AspectRatio != "9:16" && req.AspectRatio != "1:1" {
		return fmt.Errorf("aspect_ratio must be one of: 16:9, 9:16, 1:1")
	}

	return nil
}

// validateScript validates a generated or updated script
func (s *ParserService) validateScript(script *domain.Script) error {
	if script.ScriptID == "" {
		return fmt.Errorf("script_id is required")
	}

	if script.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if len(script.Scenes) == 0 {
		return fmt.Errorf("script must have at least one scene")
	}

	// Validate scene timing consistency
	var totalDuration float64
	for i, scene := range script.Scenes {
		if scene.SceneNumber != i+1 {
			return fmt.Errorf("scene %d has incorrect scene_number %d", i+1, scene.SceneNumber)
		}

		if scene.Duration <= 0 {
			return fmt.Errorf("scene %d has invalid duration %.2f", scene.SceneNumber, scene.Duration)
		}

		if scene.GenerationPrompt == "" {
			return fmt.Errorf("scene %d missing generation_prompt", scene.SceneNumber)
		}

		totalDuration += scene.Duration
	}

	// Allow slight timing discrepancies (±1 second)
	if totalDuration < float64(script.TotalDuration)-1 || totalDuration > float64(script.TotalDuration)+1 {
		s.logger.Warn("Scene timing mismatch",
			zap.Float64("calculated_duration", totalDuration),
			zap.Int("declared_duration", script.TotalDuration))
	}

	return nil
}
