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

// VeoAdapter implements VideoGeneratorAdapter for Google Veo 3.1
type VeoAdapter struct {
	apiToken     string
	httpClient   *http.Client
	logger       *zap.Logger
	modelVersion string
}

// NewVeoAdapter creates a new Veo 3.1 adapter
func NewVeoAdapter(apiToken string, logger *zap.Logger) *VeoAdapter {
	return &VeoAdapter{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // Async operation - just for initial request acknowledgment
		},
		logger:       logger,
		// Veo 3.1 model on Replicate with specific version hash
		// Replicate HTTP API requires the full version hash for direct API calls
		// Version hash obtained from: https://replicate.com/google/veo-3.1
		modelVersion: "google/veo-3.1:a55204f92195a6c535170095e221116968f43614517d8ad32b338fa12ee4460b",
	}
}

// VeoRequest matches the Veo 3.1 API schema on Replicate
type VeoRequest struct {
	Version string                 `json:"version"`
	Input   map[string]interface{} `json:"input"`
}

// VeoResponse represents the Replicate API response
type VeoResponse struct {
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

// GenerateVideo submits a video generation request to Veo 3.1
func (v *VeoAdapter) GenerateVideo(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	v.logger.Info("Generating video with Veo 3.1",
		zap.String("prompt", req.Prompt),
		zap.Int("duration", req.Duration),
		zap.String("aspect_ratio", req.AspectRatio),
	)

	// Map our aspect ratio to Veo's format
	aspectRatio := v.mapAspectRatio(req.AspectRatio)

	// Build the full prompt with style
	fullPrompt := req.Prompt
	if req.Style != "" {
		fullPrompt = fmt.Sprintf("%s. Style: %s", req.Prompt, req.Style)
	}

	// Construct Veo API request input
	// Veo 3.1 API schema: prompt, aspect_ratio, duration, image, last_frame, reference_images, negative_prompt, resolution
	input := map[string]interface{}{
		"prompt":       fullPrompt,
		"aspect_ratio": aspectRatio,
		"duration":     v.mapDuration(req.Duration),
	}

	// Add image if provided (Veo uses "image" not "start_image")
	if req.StartImageURL != "" {
		input["image"] = req.StartImageURL
		v.logger.Info("Using start image for video generation",
			zap.String("image_url", req.StartImageURL),
		)
	}

	// Add negative_prompt if provided
	if req.NegativePrompt != "" {
		input["negative_prompt"] = req.NegativePrompt
	}

	// Note: Veo 3.1 also supports:
	// - last_frame: for interpolation between two images
	// - reference_images: array of 1-3 reference images (only works with 16:9 and 8s duration)
	// - resolution: optional resolution setting
	// These can be added later if needed

	// Construct Veo API request
	// Replicate API format: either "model-name" or "model-name:version-hash"
	// If modelVersion doesn't have a colon, Replicate will use the latest version
	veoReq := VeoRequest{
		Version: v.modelVersion,
		Input:   input,
	}
	
	// Log the full request for debugging
	v.logger.Debug("Veo API request",
		zap.String("version", v.modelVersion),
		zap.Any("input", input),
	)

	// Marshal request
	jsonData, err := json.Marshal(veoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	v.logger.Info("Submitting Veo API request",
		zap.Int("clip_duration_seconds", v.mapDuration(req.Duration)),
		zap.String("aspect_ratio", aspectRatio),
		zap.Bool("has_image", req.StartImageURL != ""),
		zap.String("model_version", v.modelVersion),
	)
	
	// Log the full request payload for debugging
	v.logger.Debug("Veo API request payload",
		zap.Any("input", input),
	)

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

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", v.apiToken))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Prefer", "wait=0") // Don't wait for completion

	// Execute request
	resp, err := v.httpClient.Do(httpReq)
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
		// Log full error details for debugging
		errorBody := string(body)
		v.logger.Error("Veo API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", errorBody),
			zap.String("request_url", httpReq.URL.String()),
			zap.String("model_version", v.modelVersion),
		)
		
		// Log the request payload for debugging 422 errors
		if resp.StatusCode == 422 {
			v.logger.Error("Veo API validation error - request payload",
				zap.Any("request_input", veoReq.Input),
				zap.String("full_request", string(jsonData)),
		)
		}
		
		// Provide specific error messages for common status codes
		if resp.StatusCode == 422 {
			// HTTP 422 - Unprocessable Entity (validation error)
			// This usually means invalid model version or parameter mismatch
			return nil, fmt.Errorf("API error (status %d): Invalid request parameters or model version. Check that model version '%s' is correct and parameters match Veo 3.1 schema. Response: %s", resp.StatusCode, v.modelVersion, errorBody)
		}
		if resp.StatusCode == 402 {
			return nil, fmt.Errorf("API error (status %d): Payment required - Replicate account has insufficient credits. Response: %s", resp.StatusCode, errorBody)
		}
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("API error (status %d): Model not found. Check that model version '%s' exists on Replicate. Response: %s", resp.StatusCode, v.modelVersion, errorBody)
		}
		
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, errorBody)
	}

	// Parse response
	var veoResp VeoResponse
	if err := json.Unmarshal(body, &veoResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	v.logger.Info("Veo prediction created successfully",
		zap.String("prediction_id", veoResp.ID),
		zap.String("status", veoResp.Status),
		zap.String("created_at", veoResp.CreatedAt),
	)

	// Map to our result format
	result := &VideoGenerationResult{
		PredictionID: veoResp.ID,
		Status:       v.mapStatus(veoResp.Status),
	}

	// If already completed (unlikely but possible)
	if veoResp.Status == "succeeded" && veoResp.Output != nil {
		if videoURL, ok := v.extractVideoURL(veoResp.Output); ok {
			result.VideoURL = videoURL
			result.Status = "completed"
		}
	}

	if veoResp.Error != "" {
		result.Status = "failed"
		result.Error = veoResp.Error
	}

	return result, nil
}

// GetStatus checks the status of a video generation job
func (v *VeoAdapter) GetStatus(ctx context.Context, predictionID string) (*VideoGenerationResult, error) {
	// Create HTTP request
	url := fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", predictionID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", v.apiToken))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := v.httpClient.Do(httpReq)
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
	var veoResp VeoResponse
	if err := json.Unmarshal(body, &veoResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Map to our result format
	result := &VideoGenerationResult{
		PredictionID: veoResp.ID,
		Status:       v.mapStatus(veoResp.Status),
	}

	// Extract video URL if completed
	if veoResp.Status == "succeeded" && veoResp.Output != nil {
		if videoURL, ok := v.extractVideoURL(veoResp.Output); ok {
			result.VideoURL = videoURL
			result.Status = "completed"
		}
	}

	// Handle failed status - check both Error field and status
	if veoResp.Status == "failed" || veoResp.Status == "canceled" {
		result.Status = "failed"
		if veoResp.Error != "" {
			result.Error = veoResp.Error
		} else if veoResp.Logs != "" {
			// If no error but logs exist, use logs as error message
			// Truncate logs if too long
			logs := veoResp.Logs
			if len(logs) > 500 {
				logs = logs[:500] + "..."
			}
			result.Error = fmt.Sprintf("Generation failed (status: %s). Logs: %s", veoResp.Status, logs)
		} else {
			// No error or logs - provide generic message with status
			result.Error = fmt.Sprintf("Generation failed with status: %s (no error details provided)", veoResp.Status)
		}
		
		// Log full response for debugging
		v.logger.Error("Veo generation failed",
			zap.String("prediction_id", veoResp.ID),
			zap.String("status", veoResp.Status),
			zap.String("error", veoResp.Error),
			zap.String("logs", veoResp.Logs),
			zap.Any("output", veoResp.Output),
		)
	}

	return result, nil
}

// GetModelName returns the name of the model
func (v *VeoAdapter) GetModelName() string {
	return "Google Veo 3.1"
}

// GetCostPerSecond returns the approximate cost per second of video
// Based on Replicate pricing for Veo 3.1
func (v *VeoAdapter) GetCostPerSecond() float64 {
	// Veo 3.1 pricing on Replicate - update this based on actual pricing
	// Estimated: ~$0.07 per second
	return 0.07
}

// mapAspectRatio maps our aspect ratio format to Veo's format
func (v *VeoAdapter) mapAspectRatio(ar string) string {
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

// mapDuration maps our duration to Veo's duration
// Veo 3.1 default duration is 8 seconds
// According to the API, duration should be an integer (default: 8)
func (v *VeoAdapter) mapDuration(seconds int) int {
	// Veo 3.1 supports various durations, default is 8 seconds
	// For scene clips, we'll use 8 seconds as the base (matches default)
	// For longer videos, we'll need to generate multiple clips
	if seconds <= 8 {
		return 8
	}
	// For longer clips, round to nearest 8-second increment
	return ((seconds + 4) / 8) * 8
}

// mapStatus maps Replicate status to our status format
func (v *VeoAdapter) mapStatus(status string) string {
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

// extractVideoURL extracts the video URL from Veo's output
func (v *VeoAdapter) extractVideoURL(output interface{}) (string, bool) {
	// Veo returns output as a string URL or array of URLs
	switch val := output.(type) {
	case string:
		return val, true
	case []interface{}:
		if len(val) > 0 {
			if url, ok := val[0].(string); ok {
				return url, true
			}
		}
	}
	return "", false
}
