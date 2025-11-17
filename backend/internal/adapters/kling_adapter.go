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
)

// KlingAdapter implements VideoGeneratorAdapter for Kling v2.5 Turbo
type KlingAdapter struct {
	apiToken     string
	httpClient   *http.Client
	logger       *zap.Logger
	modelVersion string
}

// NewKlingAdapter creates a new Kling adapter
func NewKlingAdapter(apiToken string, logger *zap.Logger) *KlingAdapter {
	return &KlingAdapter{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:       logger,
		modelVersion: "kwaivgi/kling-v2.5-turbo-pro:939cd1851c5b112f284681b57ee9b0f36d0f913ba97de5845a7eef92d52837df",
	}
}

// KlingRequest matches the Kling v2.5 Turbo API schema
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

// GenerateVideo submits a video generation request to Kling v2.5 Turbo
func (k *KlingAdapter) GenerateVideo(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	k.logger.Info("Generating video with Kling v2.5 Turbo",
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

	// Construct Kling API request input
	input := map[string]interface{}{
		"prompt":   fullPrompt,
		"duration": k.mapDuration(req.Duration),
	}

	// Add start_image if provided (takes precedence over aspect_ratio)
	if req.StartImageURL != "" {
		input["start_image"] = req.StartImageURL
		k.logger.Info("Using start image for video generation",
			zap.String("start_image_url", req.StartImageURL),
		)
	} else {
		// Only set aspect_ratio if no start_image is provided
		input["aspect_ratio"] = aspectRatio
	}

	// Add negative_prompt if provided
	if req.NegativePrompt != "" {
		input["negative_prompt"] = req.NegativePrompt
	}

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

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.replicate.com/v1/predictions",
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", k.apiToken))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Prefer", "wait=0") // Don't wait for completion

	// Execute request
	resp, err := k.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		k.logger.Error("Kling API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var klingResp KlingResponse
	if err := json.Unmarshal(body, &klingResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	k.logger.Info("Video generation started",
		zap.String("prediction_id", klingResp.ID),
		zap.String("status", klingResp.Status),
	)

	// Map to our result format
	result := &VideoGenerationResult{
		PredictionID: klingResp.ID,
		Status:       k.mapStatus(klingResp.Status),
	}

	// If already completed (unlikely but possible)
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
	// Create HTTP request
	url := fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", predictionID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", k.apiToken))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := k.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var klingResp KlingResponse
	if err := json.Unmarshal(body, &klingResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Map to our result format
	result := &VideoGenerationResult{
		PredictionID: klingResp.ID,
		Status:       k.mapStatus(klingResp.Status),
	}

	// Extract video URL if completed
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

// GetModelName returns the name of the model
func (k *KlingAdapter) GetModelName() string {
	return "Kling v2.5 Turbo Pro"
}

// GetCostPerSecond returns the approximate cost per second of video
// Based on Replicate pricing for Kling v2.5 Turbo
func (k *KlingAdapter) GetCostPerSecond() float64 {
	// Official Replicate pricing: $0.07 per second
	// For a 30s video: $2.10 (slightly over target but acceptable for MVP)
	// For a 60s video: $4.20 (we'll need to optimize with parallel clip generation)
	return 0.07
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
		// Default to 16:9
		return "16:9"
	}
}

// mapDuration maps our duration to Kling's duration
// Kling v2.5 Turbo supports 5s or 10s clips
func (k *KlingAdapter) mapDuration(seconds int) string {
	// For longer videos, we'll need to generate multiple clips
	// For now, use 10s as the base clip duration
	if seconds <= 5 {
		return "5"
	}
	return "10"
}

// mapStatus maps Replicate status to our status format
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
	// Kling returns output as a string URL or array of URLs
	switch v := output.(type) {
	case string:
		return v, true
	case []interface{}:
		if len(v) > 0 {
			if url, ok := v[0].(string); ok {
				return url, true
			}
		}
	}
	return "", false
}
