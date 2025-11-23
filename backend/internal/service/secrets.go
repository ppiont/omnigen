package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.uber.org/zap"
)

// SecretsService handles Secrets Manager operations
type SecretsService struct {
	client             *secretsmanager.Client
	replicateSecretARN string
	openaiSecretARN    string
	logger             *zap.Logger
}

// NewSecretsService creates a new Secrets Manager service
func NewSecretsService(
	client *secretsmanager.Client,
	replicateSecretARN string,
	openaiSecretARN string,
	logger *zap.Logger,
) *SecretsService {
	return &SecretsService{
		client:             client,
		replicateSecretARN: replicateSecretARN,
		openaiSecretARN:    openaiSecretARN,
		logger:             logger,
	}
}

// APIKeysSecret represents the structure of API keys in Secrets Manager
type APIKeysSecret struct {
	APIKeys []string `json:"api_keys"`
}

// GetAPIKeys retrieves API keys from Secrets Manager
func (s *SecretsService) GetAPIKeys(ctx context.Context) ([]string, error) {
	// For MVP, we'll use a hardcoded secret name for API keys
	// In production, this should be configurable
	secretName := "omnigen/api-keys"

	s.logger.Info("Retrieving API keys from Secrets Manager",
		zap.String("secret_name", secretName),
	)

	result, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		s.logger.Warn("Failed to retrieve API keys from Secrets Manager, using default",
			zap.Error(err),
		)
		// Fallback to default API key for development
		return []string{"dev-api-key-12345"}, nil
	}

	var secret APIKeysSecret
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		s.logger.Error("Failed to unmarshal secret", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	s.logger.Info("API keys retrieved successfully", zap.Int("count", len(secret.APIKeys)))
	return secret.APIKeys, nil
}

// GetReplicateAPIKey retrieves the Replicate API key
func (s *SecretsService) GetReplicateAPIKey(ctx context.Context) (string, error) {
	// Check environment variable first (for local development)
	if apiKey := os.Getenv("REPLICATE_API_KEY"); apiKey != "" {
		s.logger.Info("Using Replicate API key from environment variable")
		return apiKey, nil
	}

	// If no secret ARN is configured, return error (can't use Secrets Manager)
	if s.replicateSecretARN == "" {
		return "", fmt.Errorf("REPLICATE_API_KEY environment variable not set and REPLICATE_SECRET_ARN not configured")
	}

	s.logger.Info("Retrieving Replicate API key from Secrets Manager",
		zap.String("secret_arn", s.replicateSecretARN),
	)

	result, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(s.replicateSecretARN),
	})
	if err != nil {
		s.logger.Error("Failed to retrieve Replicate API key", zap.Error(err))
		return "", fmt.Errorf("failed to retrieve replicate API key: %w", err)
	}

	return *result.SecretString, nil
}

// GetOpenAIAPIKey retrieves the OpenAI API key
func (s *SecretsService) GetOpenAIAPIKey(ctx context.Context) (string, error) {
	// Check environment variable first (for local development)
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		s.logger.Info("Using OpenAI API key from environment variable")
		return apiKey, nil
	}

	// If no secret ARN is configured, return error (can't use Secrets Manager)
	if s.openaiSecretARN == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set and OPENAI_SECRET_ARN not configured")
	}

	s.logger.Info("Retrieving OpenAI API key from Secrets Manager",
		zap.String("secret_arn", s.openaiSecretARN),
	)

	result, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(s.openaiSecretARN),
	})
	if err != nil {
		s.logger.Error("Failed to retrieve OpenAI API key", zap.Error(err))
		return "", fmt.Errorf("failed to retrieve OpenAI API key: %w", err)
	}

	return *result.SecretString, nil
}
