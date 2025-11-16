package service

import (
	"context"
	"encoding/json"
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
	"github.com/omnigen/backend/internal/prompts"
)

// ParserService handles ad script generation and management
type ParserService struct {
	llama        *adapters.LlamaAdapter
	dynamoDB     *dynamodb.Client
	scriptsTable string
	logger       *zap.Logger
}

// ParseRequest represents user input for script generation
type ParseRequest struct {
	UserID         string `json:"user_id"`
	Prompt         string `json:"prompt"`   // Free-form user input
	Duration       int    `json:"duration"` // 15, 30, or 60 seconds
	ProductName    string `json:"product_name"`
	TargetAudience string `json:"target_audience"`
	BrandVibe      string `json:"brand_vibe,omitempty"` // Optional style guidance
}

// NewParserService creates a new script parser service
func NewParserService(
	llama *adapters.LlamaAdapter,
	dynamoDB *dynamodb.Client,
	scriptsTable string,
	logger *zap.Logger,
) *ParserService {
	return &ParserService{
		llama:        llama,
		dynamoDB:     dynamoDB,
		scriptsTable: scriptsTable,
		logger:       logger,
	}
}

// GenerateScript creates a new ad script using LLM
func (s *ParserService) GenerateScript(ctx context.Context, req ParseRequest) (*domain.Script, error) {
	s.logger.Info("Generating script",
		zap.String("user_id", req.UserID),
		zap.Int("duration", req.Duration),
		zap.String("product", req.ProductName))

	// Validate request
	if err := s.validateParseRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Construct user prompt from request
	userPrompt := s.buildUserPrompt(req)

	// Call LLM with system prompt + user prompt
	llmReq := adapters.LlamaRequest{
		SystemPrompt: prompts.AdScriptSystemPrompt + "\n\n" + prompts.AdScriptFewShotExamples,
		UserPrompt:   userPrompt,
		Temperature:  0.8,  // Creative but controlled
		MaxTokens:    8000, // Enough for detailed scripts
	}

	output, err := s.llama.GenerateScript(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Extract JSON from output (handles markdown code blocks)
	jsonStr, err := adapters.ExtractJSON(output)
	if err != nil {
		s.logger.Error("Failed to extract JSON from LLM output",
			zap.String("output", output),
			zap.Error(err))
		return nil, fmt.Errorf("failed to extract JSON from LLM output: %w", err)
	}

	// Parse the generated script
	var scriptData struct {
		Title         string           `json:"title"`
		TotalDuration int              `json:"total_duration"`
		Scenes        []domain.Scene   `json:"scenes"`
		AudioSpec     domain.AudioSpec `json:"audio_spec"`
		Metadata      domain.Metadata  `json:"metadata"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &scriptData); err != nil {
		s.logger.Error("Failed to parse LLM JSON output",
			zap.String("json", jsonStr),
			zap.Error(err))
		return nil, fmt.Errorf("failed to parse script JSON: %w", err)
	}

	// Create script domain object
	script := &domain.Script{
		ScriptID:      fmt.Sprintf("script-%s", uuid.New().String()),
		UserID:        req.UserID,
		Title:         scriptData.Title,
		TotalDuration: scriptData.TotalDuration,
		Scenes:        scriptData.Scenes,
		AudioSpec:     scriptData.AudioSpec,
		Metadata:      scriptData.Metadata,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
		Status:        "draft",
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour).Unix(), // 30-day TTL
	}

	// Validate the generated script
	if err := s.validateScript(script); err != nil {
		return nil, fmt.Errorf("generated script validation failed: %w", err)
	}

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

// buildUserPrompt constructs the user prompt from the parse request
func (s *ParserService) buildUserPrompt(req ParseRequest) string {
	prompt := fmt.Sprintf(`Create a %d-second advertisement script with the following requirements:

**Product**: %s
**Target Audience**: %s
**User Description**: %s`,
		req.Duration,
		req.ProductName,
		req.TargetAudience,
		req.Prompt)

	if req.BrandVibe != "" {
		prompt += fmt.Sprintf("\n**Brand Vibe/Style**: %s", req.BrandVibe)
	}

	prompt += "\n\nGenerate a production-ready script with industry-standard cinematography terminology."

	return prompt
}

// validateParseRequest validates the parse request
func (s *ParserService) validateParseRequest(req ParseRequest) error {
	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	if req.Duration != 15 && req.Duration != 30 && req.Duration != 60 {
		return fmt.Errorf("duration must be 15, 30, or 60 seconds")
	}

	if req.ProductName == "" {
		return fmt.Errorf("product_name is required")
	}

	if req.TargetAudience == "" {
		return fmt.Errorf("target_audience is required")
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
