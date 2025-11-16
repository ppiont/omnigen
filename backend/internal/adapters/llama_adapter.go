package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.uber.org/zap"
)

// LlamaAdapter handles LLM interactions with Llama 3.1 405B via Replicate
type LlamaAdapter struct {
	apiToken      string
	httpClient    *http.Client
	logger        *zap.Logger
	secretManager *secretsmanager.Client
	secretARN     string
}

// LlamaRequest represents a request to Llama via Replicate
type LlamaRequest struct {
	SystemPrompt string  `json:"system_prompt"`
	UserPrompt   string  `json:"user_prompt"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"max_tokens"`
}

// LlamaResponse represents the parsed response from Llama
type LlamaResponse struct {
	Output string `json:"output"`
}

// NewLlamaAdapter creates a new Llama adapter instance
func NewLlamaAdapter(secretARN string, sm *secretsmanager.Client, logger *zap.Logger) *LlamaAdapter {
	return &LlamaAdapter{
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // LLM calls can be slow
		},
		logger:        logger,
		secretManager: sm,
		secretARN:     secretARN,
	}
}

// Initialize loads the Replicate API token from environment or Secrets Manager
func (a *LlamaAdapter) Initialize(ctx context.Context) error {
	// Check for environment variable first (for local development)
	if apiKey := os.Getenv("REPLICATE_API_KEY"); apiKey != "" {
		a.apiToken = apiKey
		a.logger.Info("Llama adapter initialized with environment variable")
		return nil
	}

	// Fall back to Secrets Manager (for production)
	output, err := a.secretManager.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &a.secretARN,
	})
	if err != nil {
		return fmt.Errorf("failed to get Replicate API token: %w", err)
	}

	// Assume secret is stored as plain text token
	a.apiToken = *output.SecretString

	a.logger.Info("Llama adapter initialized successfully from Secrets Manager")
	return nil
}

// GenerateScript calls Llama to generate a structured ad script
func (a *LlamaAdapter) GenerateScript(ctx context.Context, req LlamaRequest) (string, error) {
	if a.apiToken == "" {
		return "", fmt.Errorf("Llama adapter not initialized - call Initialize() first")
	}

	a.logger.Info("Generating script with Llama",
		zap.Int("user_prompt_length", len(req.UserPrompt)),
		zap.Float64("temperature", req.Temperature))

	// Create Replicate prediction
	predictionID, err := a.createPrediction(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create prediction: %w", err)
	}

	a.logger.Info("Prediction created", zap.String("prediction_id", predictionID))

	// Poll for completion
	output, err := a.pollPrediction(ctx, predictionID)
	if err != nil {
		return "", fmt.Errorf("failed to get prediction result: %w", err)
	}

	a.logger.Info("Script generation completed",
		zap.String("prediction_id", predictionID),
		zap.Int("output_length", len(output)))

	return output, nil
}

// createPrediction starts a new Llama inference on Replicate
func (a *LlamaAdapter) createPrediction(ctx context.Context, req LlamaRequest) (string, error) {
	// Replicate API endpoint for Llama 3.1 405B Instruct
	url := "https://api.replicate.com/v1/predictions"

	// Construct the full prompt (system + user)
	fullPrompt := req.SystemPrompt + "\n\n" + req.UserPrompt

	// Replicate prediction request body
	// Using owner/name format (Replicate will use latest version)
	body := map[string]interface{}{
		"version": "meta/meta-llama-3.1-405b-instruct",
		"input": map[string]interface{}{
			"prompt":      fullPrompt,
			"temperature": req.Temperature,
			"max_tokens":  req.MaxTokens,
			"top_p":       0.9,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Token "+a.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ID, nil
}

// pollPrediction polls Replicate until the prediction completes
func (a *LlamaAdapter) pollPrediction(ctx context.Context, predictionID string) (string, error) {
	url := fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", predictionID)

	// Poll every 2 seconds for up to 2 minutes
	maxAttempts := 60
	pollInterval := 2 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(pollInterval):
			// Continue to poll
		}

		httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create poll request: %w", err)
		}

		httpReq.Header.Set("Authorization", "Token "+a.apiToken)

		resp, err := a.httpClient.Do(httpReq)
		if err != nil {
			return "", fmt.Errorf("poll request failed: %w", err)
		}

		var result struct {
			Status string   `json:"status"`
			Output []string `json:"output"`
			Error  string   `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return "", fmt.Errorf("failed to decode poll response: %w", err)
		}
		resp.Body.Close()

		a.logger.Debug("Poll attempt",
			zap.Int("attempt", attempt+1),
			zap.String("status", result.Status))

		switch result.Status {
		case "succeeded":
			if len(result.Output) == 0 {
				return "", fmt.Errorf("no output from model")
			}
			// Join all output chunks (Replicate streams output as array)
			return strings.Join(result.Output, ""), nil

		case "failed", "canceled":
			return "", fmt.Errorf("prediction failed: %s", result.Error)

		case "starting", "processing":
			// Continue polling
			continue

		default:
			a.logger.Warn("Unknown prediction status", zap.String("status", result.Status))
			continue
		}
	}

	return "", fmt.Errorf("prediction timed out after %d attempts", maxAttempts)
}

// ExtractJSON extracts JSON from LLM output (handles markdown code blocks)
func ExtractJSON(output string) (string, error) {
	// Try to find JSON in markdown code block
	if strings.Contains(output, "```json") {
		start := strings.Index(output, "```json") + 7
		end := strings.Index(output[start:], "```")
		if end != -1 {
			return strings.TrimSpace(output[start : start+end]), nil
		}
	}

	// Try to find JSON in generic code block
	if strings.Contains(output, "```") {
		start := strings.Index(output, "```") + 3
		end := strings.Index(output[start:], "```")
		if end != -1 {
			extracted := strings.TrimSpace(output[start : start+end])
			// Verify it looks like JSON
			if strings.HasPrefix(extracted, "{") {
				return extracted, nil
			}
		}
	}

	// Assume entire output is JSON
	trimmed := strings.TrimSpace(output)
	if strings.HasPrefix(trimmed, "{") {
		return trimmed, nil
	}

	return "", fmt.Errorf("could not extract JSON from output")
}
