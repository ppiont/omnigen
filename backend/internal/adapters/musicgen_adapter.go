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

// MusicGenAdapter implements music generation using Meta's MusicGen
type MusicGenAdapter struct {
	apiToken     string
	httpClient   *http.Client
	logger       *zap.Logger
	modelVersion string
}

// NewMusicGenAdapter creates a new MusicGen adapter
func NewMusicGenAdapter(apiToken string, logger *zap.Logger) *MusicGenAdapter {
	return &MusicGenAdapter{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:       logger,
		modelVersion: "meta/musicgen:latest",
	}
}

// MusicGenRequest represents a music generation request
type MusicGenRequest struct {
	Prompt   string `json:"prompt"`
	Duration int    `json:"duration"` // in seconds
	Model    string `json:"model"`    // melody, small, medium, large
}

// MusicGenResponse represents the Replicate API response
type MusicGenResponse struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	Output      interface{}            `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	CompletedAt string                 `json:"completed_at,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
}

// AudioGenerationResult represents the result of audio generation
type AudioGenerationResult struct {
	AudioURL     string
	PredictionID string
	Status       string // "processing", "completed", "failed"
	Error        string
}

// GenerateMusic submits a music generation request
func (m *MusicGenAdapter) GenerateMusic(ctx context.Context, req *MusicGenRequest) (*AudioGenerationResult, error) {
	m.logger.Info("Generating music with MusicGen",
		zap.String("prompt", req.Prompt),
		zap.Int("duration", req.Duration),
	)

	// Build the request input
	input := map[string]interface{}{
		"prompt":   req.Prompt,
		"duration": req.Duration,
	}

	if req.Model != "" {
		input["model_version"] = req.Model
	}

	// Construct request
	replicateReq := map[string]interface{}{
		"version": m.modelVersion,
		"input":   input,
	}

	// Marshal request
	jsonData, err := json.Marshal(replicateReq)
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

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.apiToken))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Prefer", "wait=0")

	// Execute request
	resp, err := m.httpClient.Do(httpReq)
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
		m.logger.Error("MusicGen API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var musicResp MusicGenResponse
	if err := json.Unmarshal(body, &musicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	m.logger.Info("Music generation started",
		zap.String("prediction_id", musicResp.ID),
		zap.String("status", musicResp.Status),
	)

	// Map to result format
	result := &AudioGenerationResult{
		PredictionID: musicResp.ID,
		Status:       mapStatus(musicResp.Status),
	}

	// If already completed (unlikely but possible)
	if musicResp.Status == "succeeded" && musicResp.Output != nil {
		if audioURL, ok := extractAudioURL(musicResp.Output); ok {
			result.AudioURL = audioURL
			result.Status = "completed"
		}
	}

	if musicResp.Error != "" {
		result.Status = "failed"
		result.Error = musicResp.Error
	}

	return result, nil
}

// GetStatus checks the status of a music generation job
func (m *MusicGenAdapter) GetStatus(ctx context.Context, predictionID string) (*AudioGenerationResult, error) {
	url := fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", predictionID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.apiToken))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(httpReq)
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

	var musicResp MusicGenResponse
	if err := json.Unmarshal(body, &musicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := &AudioGenerationResult{
		PredictionID: musicResp.ID,
		Status:       mapStatus(musicResp.Status),
	}

	if musicResp.Status == "succeeded" && musicResp.Output != nil {
		if audioURL, ok := extractAudioURL(musicResp.Output); ok {
			result.AudioURL = audioURL
			result.Status = "completed"
		}
	}

	if musicResp.Error != "" {
		result.Status = "failed"
		result.Error = musicResp.Error
	}

	return result, nil
}

// Helper functions
func mapStatus(status string) string {
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

func extractAudioURL(output interface{}) (string, bool) {
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
