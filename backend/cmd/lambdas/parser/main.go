package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.uber.org/zap"

	"github.com/omnigen/backend/internal/adapters"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/prompts"
)

// Input represents the Lambda input for script generation
type Input struct {
	ScriptID       string `json:"script_id"`
	UserID         string `json:"user_id"`
	Prompt         string `json:"prompt"`
	Duration       int    `json:"duration"`
	ProductName    string `json:"product_name"`
	TargetAudience string `json:"target_audience"`
	BrandVibe      string `json:"brand_vibe,omitempty"`
	Style          string `json:"style,omitempty"`
}

var (
	dynamoClient  *dynamodb.Client
	secretsClient *secretsmanager.Client
	llamaAdapter  *adapters.LlamaAdapter
	scriptsTable  string
	secretARN     string
	logger        *zap.Logger
)

func init() {
	// Initialize logger
	logger, _ = zap.NewProduction()

	scriptsTable = os.Getenv("SCRIPTS_TABLE")
	if scriptsTable == "" {
		log.Fatal("SCRIPTS_TABLE environment variable not set")
	}

	secretARN = os.Getenv("REPLICATE_SECRET_ARN")
	if secretARN == "" {
		log.Fatal("REPLICATE_SECRET_ARN environment variable not set")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	secretsClient = secretsmanager.NewFromConfig(cfg)

	// Initialize LlamaAdapter with Secrets Manager client and logger
	llamaAdapter = adapters.NewLlamaAdapter(secretARN, secretsClient, logger)
}

func handler(ctx context.Context, input Input) error {
	log.Printf("Parser Lambda invoked for script %s (user: %s)", input.ScriptID, input.UserID)

	// Initialize Llama adapter if not already initialized
	if err := llamaAdapter.Initialize(ctx); err != nil {
		log.Printf("Failed to initialize Llama adapter: %v", err)
		return updateScriptStatus(ctx, input.ScriptID, "failed", fmt.Sprintf("Failed to initialize LLM: %v", err))
	}

	// Validate input
	if err := validateInput(input); err != nil {
		log.Printf("Invalid input: %v", err)
		return updateScriptStatus(ctx, input.ScriptID, "failed", fmt.Sprintf("Invalid input: %v", err))
	}

	// Build user prompt from input
	userPrompt := buildUserPrompt(input)
	log.Printf("Calling LLM with prompt for %s (%ds duration)", input.ProductName, input.Duration)

	// Call LLM with system prompt + user prompt
	llmReq := adapters.LlamaRequest{
		SystemPrompt: prompts.AdScriptSystemPrompt + "\n\n" + prompts.AdScriptFewShotExamples,
		UserPrompt:   userPrompt,
		Temperature:  0.8,  // Creative but controlled
		MaxTokens:    8000, // Enough for detailed scripts
	}

	output, err := llamaAdapter.GenerateScript(ctx, llmReq)
	if err != nil {
		log.Printf("LLM generation failed: %v", err)
		return updateScriptStatus(ctx, input.ScriptID, "failed", fmt.Sprintf("LLM generation failed: %v", err))
	}

	log.Printf("LLM generation completed, extracting JSON")

	// Extract JSON from output (handles markdown code blocks)
	jsonStr, err := adapters.ExtractJSON(output)
	if err != nil {
		log.Printf("Failed to extract JSON from LLM output: %v", err)
		return updateScriptStatus(ctx, input.ScriptID, "failed", fmt.Sprintf("Failed to extract JSON: %v", err))
	}

	// Log the extracted JSON for debugging
	log.Printf("Extracted JSON length: %d bytes", len(jsonStr))

	// Write full JSON to CloudWatch for inspection (will be truncated if too long)
	if len(jsonStr) > 5000 {
		log.Printf("Full JSON (truncated): %s...", jsonStr[:5000])
	} else {
		log.Printf("Full JSON: %s", jsonStr)
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
		log.Printf("Failed to parse LLM JSON output: %v", err)
		return updateScriptStatus(ctx, input.ScriptID, "failed", fmt.Sprintf("Failed to parse JSON: %v", err))
	}

	log.Printf("Successfully parsed script: %s with %d scenes", scriptData.Title, len(scriptData.Scenes))

	// Create complete script domain object
	script := &domain.Script{
		ScriptID:      input.ScriptID,
		UserID:        input.UserID,
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
	if err := validateScript(script); err != nil {
		log.Printf("Generated script validation failed: %v", err)
		return updateScriptStatus(ctx, input.ScriptID, "failed", fmt.Sprintf("Validation failed: %v", err))
	}

	// Save complete script to DynamoDB
	if err := saveScript(ctx, script); err != nil {
		log.Printf("Failed to save script: %v", err)
		return updateScriptStatus(ctx, input.ScriptID, "failed", fmt.Sprintf("Failed to save: %v", err))
	}

	log.Printf("Successfully generated and saved script %s", input.ScriptID)
	return nil
}

// getReplicateAPIKey fetches the API key from AWS Secrets Manager
func getReplicateAPIKey(ctx context.Context) (string, error) {
	result, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretARN),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	var secretData struct {
		APIKey string `json:"api_key"`
	}

	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		return "", fmt.Errorf("failed to parse secret: %w", err)
	}

	return secretData.APIKey, nil
}

// buildUserPrompt constructs the user prompt from input
func buildUserPrompt(input Input) string {
	prompt := fmt.Sprintf(`Create a %d-second advertisement script with the following requirements:

**Product**: %s
**Target Audience**: %s
**User Description**: %s`,
		input.Duration,
		input.ProductName,
		input.TargetAudience,
		input.Prompt)

	if input.BrandVibe != "" {
		prompt += fmt.Sprintf("\n**Brand Vibe/Style**: %s", input.BrandVibe)
	}

	prompt += "\n\nGenerate a production-ready script with industry-standard cinematography terminology."

	return prompt
}

// validateInput validates the Lambda input
func validateInput(input Input) error {
	if input.ScriptID == "" {
		return fmt.Errorf("script_id is required")
	}
	if input.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if input.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if input.Duration != 15 && input.Duration != 30 && input.Duration != 60 {
		return fmt.Errorf("duration must be 15, 30, or 60 seconds")
	}
	if input.ProductName == "" {
		return fmt.Errorf("product_name is required")
	}
	if input.TargetAudience == "" {
		return fmt.Errorf("target_audience is required")
	}
	return nil
}

// validateScript validates a generated script
func validateScript(script *domain.Script) error {
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
		log.Printf("Warning: Scene timing mismatch - calculated: %.2f, declared: %d", totalDuration, script.TotalDuration)
	}

	return nil
}

// saveScript saves the complete script to DynamoDB
func saveScript(ctx context.Context, script *domain.Script) error {
	item, err := attributevalue.MarshalMap(script)
	if err != nil {
		return fmt.Errorf("failed to marshal script: %w", err)
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(scriptsTable),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to save script to DynamoDB: %w", err)
	}

	log.Printf("Script saved to SCRIPTS_TABLE: %s", script.ScriptID)
	return nil
}

// updateScriptStatus updates the script status (used for error handling)
func updateScriptStatus(ctx context.Context, scriptID, status, message string) error {
	now := time.Now()

	updateExpr := "SET #status = :status, updated_at = :updated_at"
	exprAttrNames := map[string]string{
		"#status": "status",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":status":     &types.AttributeValueMemberS{Value: status},
		":updated_at": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now.Unix())},
	}

	// If there's an error message, include it
	if message != "" {
		updateExpr += ", error_message = :error"
		exprAttrValues[":error"] = &types.AttributeValueMemberS{Value: message}
	}

	_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(scriptsTable),
		Key: map[string]types.AttributeValue{
			"script_id": &types.AttributeValueMemberS{Value: scriptID},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: exprAttrValues,
	})

	if err != nil {
		log.Printf("Failed to update script status: %v", err)
		return err
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
