package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
			Timeout: 30 * time.Second, // Async operation - just for initial request
		},
		logger:       logger,
		modelVersion: "google/veo-3.1:20ebd92c5919f20e8fa2e983bdb60016a99794c9accfab496ea25a68e0dbbaad",
	}
}

// VeoRequest represents the Replicate API request for Veo 3.1
type VeoRequest struct {
	Version string                 `json:"version"`
	Input   map[string]interface{} `json:"input"`
}

// VeoResponse represents the Replicate API response
type VeoResponse struct {
	ID          string      `json:"id"`
	Status      string      `json:"status"`
	Output      interface{} `json:"output,omitempty"`
	Error       string      `json:"error,omitempty"`
	Logs        string      `json:"logs,omitempty"`
	CreatedAt   string      `json:"created_at"`
	CompletedAt string      `json:"completed_at,omitempty"`
}

// GenerateVideo submits a video generation request to Veo 3.1
func (v *VeoAdapter) GenerateVideo(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error) {
	v.logger.Info("Generating video with Veo 3.1",
		zap.String("prompt", req.Prompt),
		zap.Int("duration", req.Duration),
		zap.String("aspect_ratio", req.AspectRatio),
		zap.Bool("has_start_image", req.StartImageURL != ""),
		zap.Bool("generate_audio", req.GenerateAudio),
	)

	// Map duration to Veo's constraints (4, 6, or 8 seconds)
	duration := v.mapDuration(req.Duration)

	// Map aspect ratio
	aspectRatio := v.mapAspectRatio(req.AspectRatio)

	// Build Veo API request
	input := map[string]interface{}{
		"prompt":         req.Prompt,
		"duration":       duration,
		"aspect_ratio":   aspectRatio,
		"resolution":     "1080p", // Default to highest quality
		"generate_audio": req.GenerateAudio,
	}

	// Add optional fields
	if req.StartImageURL != "" {
		input["image"] = req.StartImageURL
	}

	if req.NegativePrompt != "" {
		input["negative_prompt"] = req.NegativePrompt
	}

	// Note: reference_images and last_frame are available but not implemented initially
	// They can be added in future enhancements

	veoReq := VeoRequest{
		Version: v.modelVersion,
		Input:   input,
	}

	// Marshal request
	jsonData, err := json.Marshal(veoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	v.logger.Debug("Sending Veo request",
		zap.String("request", string(jsonData)),
	)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.replicate.com/v1/predictions",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", v.apiToken))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Prefer", "wait=0") // Async mode

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
		v.logger.Error("Veo API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var veoResp VeoResponse
	if err := json.Unmarshal(body, &veoResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	v.logger.Info("Veo prediction created",
		zap.String("prediction_id", veoResp.ID),
		zap.String("status", veoResp.Status),
	)

	// Return result
	return &VideoGenerationResult{
		PredictionID: veoResp.ID,
		Status:       v.mapStatus(veoResp.Status),
		HasAudio:     req.GenerateAudio,
	}, nil
}

// GetStatus checks the status of a video generation job
func (v *VeoAdapter) GetStatus(ctx context.Context, predictionID string) (*VideoGenerationResult, error) {
	url := fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", predictionID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", v.apiToken))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var veoResp VeoResponse
	if err := json.Unmarshal(body, &veoResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := &VideoGenerationResult{
		PredictionID: veoResp.ID,
		Status:       v.mapStatus(veoResp.Status),
		// Note: We don't know if audio was generated from status check alone
		// This would need to be tracked in the job record
	}

	// Extract video URL if completed
	if veoResp.Status == "succeeded" && veoResp.Output != nil {
		if videoURL, ok := veoResp.Output.(string); ok {
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

// GetModelName returns the name of the model
func (v *VeoAdapter) GetModelName() string {
	return "Google Veo 3.1"
}

// GetCostPerSecond returns the cost per second of video
func (v *VeoAdapter) GetCostPerSecond() float64 {
	// This should ideally check if audio is being generated
	// For now, default to audio-enabled pricing
	return 0.15
}

// mapDuration maps requested duration to Veo's supported durations (4, 6, or 8 seconds)
func (v *VeoAdapter) mapDuration(seconds int) int {
	if seconds <= 4 {
		return 4
	} else if seconds <= 6 {
		return 6
	}
	return 8
}

// mapAspectRatio maps aspect ratio to Veo format
func (v *VeoAdapter) mapAspectRatio(ratio string) string {
	switch ratio {
	case "16:9":
		return "16:9"
	case "9:16":
		return "9:16"
	case "1:1":
		// Veo doesn't support 1:1, default to 16:9
		v.logger.Warn("Veo doesn't support 1:1 aspect ratio, defaulting to 16:9")
		return "16:9"
	default:
		return "16:9"
	}
}

// mapStatus maps Replicate status to internal format
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
