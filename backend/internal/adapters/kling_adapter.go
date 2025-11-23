package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/omnigen/backend/pkg/retry"
)

// KlingAdapter implements VideoGeneratorAdapter for Kling V2.5 Turbo Pro
type KlingAdapter struct {
	apiToken     string
	httpClient   *http.Client
	logger       *zap.Logger
	modelVersion string
}

// NewKlingAdapter creates a new Kling V2.5 adapter
func NewKlingAdapter(apiToken string, logger *zap.Logger) *KlingAdapter {
	return &KlingAdapter{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
		// Kling V2.5 Turbo Pro model on Replicate
		// Replicate HTTP API requires the full version hash for direct API calls
		// To get the version hash:
		//   1. Visit https://replicate.com/kwaivgi/kling-v2.5-turbo-pro
		//   2. Click on a version to see the hash in the URL
		//   3. Or use: curl -H "Authorization: Bearer $REPLICATE_API_TOKEN" https://api.replicate.com/v1/models/kwaivgi/kling-v2.5-turbo-pro/versions
		// Using :latest as fallback - should be replaced with specific version hash for production
		modelVersion: "kwaivgi/kling-v2.5-turbo-pro:latest",
	}
}

// KlingRequest matches the Kling API schema on Replicate
type KlingRequest struct {
	Version string                 `json:"version"`
	Input   map[string]interface{} `json:"input"`
}

// KlingResponse represents the Replicate API response
type KlingResponse struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	Output      interface{}            `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Logs        string                 `json:"logs,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	CompletedAt string                 `json:"completed_at,omitempty"`
	URLs        map[string]string      `json:"urls,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
}

// GenerateVideo submits a video generation request to Kling V2.5
func (k *KlingAdapter) GenerateVideo(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	k.logger.Info("Generating video with Kling V2.5",
		zap.String("prompt", req.Prompt),
		zap.Int("duration", req.Duration),
		zap.String("aspect_ratio", req.AspectRatio),
	)

	// Map our aspect ratio to Kling's format
	aspectRatio := k.mapAspectRatio(req.AspectRatio)

	// Build the full prompt with style
	fullPrompt := req.Prompt
	if req.Style != "" {
		fullPrompt = fmt.Sprintf("%s. Style: %s", req.Prompt, req.Style)
	}

	// Construct Kling API request input based on actual schema
	// Kling API: prompt, aspect_ratio, duration, start_image, negative_prompt, guidance_scale
	input := map[string]interface{}{
		"prompt":       fullPrompt,
		"aspect_ratio": aspectRatio,
		"duration":     k.mapDuration(req.Duration),
	}

	// Add start_image if provided (Kling uses "start_image" not "image")
	if req.StartImageURL != "" {
		input["start_image"] = req.StartImageURL
		k.logger.Info("Using start image for video generation",
			zap.String("image_url", req.StartImageURL),
		)
	}

	// Add negative_prompt if provided
	if req.NegativePrompt != "" {
		input["negative_prompt"] = req.NegativePrompt
	}

	// Optional: Add guidance_scale (default is 0.5 based on schema)
	// Can be made configurable later if needed

	// Construct Kling API request
	klingReq := KlingRequest{
		Version: k.modelVersion,
		Input:   input,
	}

	// Marshal request
	jsonData, err := json.Marshal(klingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	k.logger.Info("Submitting Kling API request",
		zap.String("model_version", k.modelVersion),
	)

	// Submit to Replicate API with retry logic
	var klingResp KlingResponse
	err = retry.Do(ctx, retry.APIConfig(), func() error {
		url := "https://api.replicate.com/v1/predictions"
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
		if err != nil {
			return retry.NewNonRetryableError(fmt.Errorf("failed to create request: %w", err))
		}

		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", k.apiToken))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Prefer", "wait=0") // Don't wait for completion

		resp, err := k.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			errorBody := string(body)
			k.logger.Error("Kling API error",
				zap.Int("status_code", resp.StatusCode),
				zap.String("response_body", errorBody),
				zap.String("request_url", httpReq.URL.String()),
				zap.String("model_version", k.modelVersion),
			)

			// Log the request payload for debugging 422 errors
			if resp.StatusCode == 422 {
				k.logger.Error("Kling API validation error - request payload",
					zap.Any("request_input", klingReq.Input),
					zap.String("full_request", string(jsonData)),
				)
			}

			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				// Provide more specific error messages for common errors
				var errMsg string
				if resp.StatusCode == 422 {
					errMsg = fmt.Sprintf("API error (status %d): Invalid request parameters or model version. Check that model version '%s' is correct and parameters match Kling V2.5 schema. Response: %s", resp.StatusCode, k.modelVersion, errorBody)
				} else if resp.StatusCode == 404 {
					errMsg = fmt.Sprintf("API error (status %d): Model not found. Check that model version '%s' exists on Replicate. Response: %s", resp.StatusCode, k.modelVersion, errorBody)
				} else {
					errMsg = fmt.Sprintf("API error: status %d, body: %s", resp.StatusCode, errorBody)
				}
				return retry.NewNonRetryableError(fmt.Errorf(errMsg))
			}
			return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, errorBody)
		}

		if err := json.Unmarshal(body, &klingResp); err != nil {
			return retry.NewNonRetryableError(fmt.Errorf("failed to parse response: %w", err))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	k.logger.Info("Kling prediction created successfully",
		zap.String("prediction_id", klingResp.ID),
		zap.String("status", klingResp.Status),
	)

	// Map to our result format
	result := &VideoGenerationResult{
		PredictionID: klingResp.ID,
		Status:       k.mapStatus(klingResp.Status),
	}

	if klingResp.Status == "succeeded" && klingResp.Output != nil {
		if videoURL, ok := k.extractVideoURL(klingResp.Output); ok {
			result.VideoURL = videoURL
			result.Status = "completed"
		}
	}

	if klingResp.Error != "" {
		result.Status = "failed"
		result.Error = klingResp.Error
	}

	return result, nil
}

// GetStatus checks the status of a video generation job
func (k *KlingAdapter) GetStatus(ctx context.Context, predictionID string) (*VideoGenerationResult, error) {
	url := fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", predictionID)

	var klingResp KlingResponse
	err := retry.Do(ctx, retry.APIConfig(), func() error {
		httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return retry.NewNonRetryableError(fmt.Errorf("failed to create request: %w", err))
		}

		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", k.apiToken))

		resp, err := k.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return retry.NewNonRetryableError(fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body)))
		}

		if err := json.Unmarshal(body, &klingResp); err != nil {
			return retry.NewNonRetryableError(fmt.Errorf("failed to parse response: %w", err))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	result := &VideoGenerationResult{
		PredictionID: klingResp.ID,
		Status:       k.mapStatus(klingResp.Status),
	}

	if klingResp.Status == "succeeded" && klingResp.Output != nil {
		if videoURL, ok := k.extractVideoURL(klingResp.Output); ok {
			result.VideoURL = videoURL
			result.Status = "completed"
		}
	}

	if klingResp.Status == "failed" || klingResp.Status == "canceled" {
		if klingResp.Error != "" {
			result.Error = klingResp.Error
		} else if klingResp.Logs != "" {
			result.Error = fmt.Sprintf("Generation failed. Logs: %s", klingResp.Logs)
		} else {
			result.Error = fmt.Sprintf("Generation failed with status: %s", klingResp.Status)
		}
		result.Status = "failed"
		k.logger.Error("Kling generation failed",
			zap.String("prediction_id", klingResp.ID),
			zap.String("status", klingResp.Status),
			zap.String("error", klingResp.Error),
			zap.String("logs", klingResp.Logs),
		)
	}

	return result, nil
}

// GetModelName returns the model name
func (k *KlingAdapter) GetModelName() string {
	return "Kling V2.5 Turbo Pro"
}

// GetCostPerSecond returns the cost per second
func (k *KlingAdapter) GetCostPerSecond() float64 {
	return 0.07 // Kling V2.5 Turbo Pro pricing: $0.07 per second
}

// mapAspectRatio maps our aspect ratio format to Kling's format
func (k *KlingAdapter) mapAspectRatio(ar string) string {
	switch ar {
	case "16:9":
		return "16:9"
	case "9:16":
		return "9:16"
	case "1:1":
		return "1:1"
	default:
		return "16:9" // Default per schema
	}
}

// mapDuration maps scene duration to Kling's supported durations
// Based on schema: duration is an integer (default 5 seconds)
// Kling typically supports 5 or 10 seconds - constrain to valid values
func (k *KlingAdapter) mapDuration(seconds int) int {
	// Kling typically supports 5 or 10 seconds
	// Map to closest valid duration
	if seconds <= 5 {
		return 5
	}
	if seconds <= 10 {
		return 10
	}
	// For longer durations, use 10 (max single clip)
	return 10
}

// mapStatus maps Kling's status to our internal status
func (k *KlingAdapter) mapStatus(status string) string {
	switch status {
	case "starting", "processing":
		return "processing"
	case "succeeded":
		return "completed"
	case "failed", "canceled":
		return "failed"
	default:
		return "processing"
	}
}

// extractVideoURL extracts the video URL from Kling's output
func (k *KlingAdapter) extractVideoURL(output interface{}) (string, bool) {
	// Kling returns output as a string URL
	if url, ok := output.(string); ok {
		return url, true
	}
	// Handle array of URLs if returned
	if urls, ok := output.([]interface{}); ok && len(urls) > 0 {
		if url, ok := urls[0].(string); ok {
			return url, true
		}
	}
	return "", false
}

