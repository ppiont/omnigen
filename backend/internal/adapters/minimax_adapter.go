package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// MinimaxAdapter implements music generation via Minimax music-1.5
type MinimaxAdapter struct {
	apiToken     string
	httpClient   *http.Client
	logger       *zap.Logger
	modelVersion string
}

// NewMinimaxAdapter creates a new Minimax music adapter
func NewMinimaxAdapter(apiToken string, logger *zap.Logger) *MinimaxAdapter {
	return &MinimaxAdapter{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:       logger,
		modelVersion: "minimax/music-1.5:latest",
	}
}

// MusicGenerationRequest represents the input for music generation
type MusicGenerationRequest struct {
	Prompt     string // User's video prompt (we'll derive music prompt from this)
	Duration   int    // Video duration in seconds
	MusicMood  string // upbeat, calm, dramatic, energetic
	MusicStyle string // electronic, acoustic, orchestral
}

// MusicGenerationResult represents the output
type MusicGenerationResult struct {
	PredictionID string
	Status       string
	AudioURL     string
	Error        string
}

// MinimaxRequest matches the Minimax API schema
type MinimaxRequest struct {
	Version string                 `json:"version"`
	Input   map[string]interface{} `json:"input"`
}

// MinimaxResponse represents the Replicate API response
type MinimaxResponse struct {
	ID          string      `json:"id"`
	Status      string      `json:"status"`
	Output      interface{} `json:"output,omitempty"`
	Error       string      `json:"error,omitempty"`
	Logs        string      `json:"logs,omitempty"`
	CreatedAt   string      `json:"created_at"`
	CompletedAt string      `json:"completed_at,omitempty"`
}

// GenerateMusic generates background music using Minimax
func (m *MinimaxAdapter) GenerateMusic(ctx context.Context, req *MusicGenerationRequest) (*MusicGenerationResult, error) {
	m.logger.Info("Generating music with Minimax",
		zap.String("mood", req.MusicMood),
		zap.String("style", req.MusicStyle),
		zap.Int("duration", req.Duration),
	)

	// Generate music prompt from video prompt and preferences
	musicPrompt := m.generateMusicPrompt(req.Prompt, req.MusicMood, req.MusicStyle)

	// Generate instrumental lyrics structure
	lyrics := m.generateLyrics(req.Duration)

	m.logger.Debug("Generated music parameters",
		zap.String("music_prompt", musicPrompt),
		zap.String("lyrics", lyrics),
	)

	// Build Minimax API request
	minimaxReq := MinimaxRequest{
		Version: m.modelVersion,
		Input: map[string]interface{}{
			"prompt":       musicPrompt,
			"lyrics":       lyrics,
			"sample_rate":  44100,
			"bitrate":      256000,
			"audio_format": "mp3",
		},
	}

	payload, err := json.Marshal(minimaxReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Submit prediction to Replicate
	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.replicate.com/v1/predictions",
		bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+m.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Prefer", "wait") // Wait for result if possible

	resp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var minimaxResp MinimaxResponse
	if err := json.Unmarshal(body, &minimaxResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := &MusicGenerationResult{
		PredictionID: minimaxResp.ID,
		Status:       minimaxResp.Status,
		Error:        minimaxResp.Error,
	}

	// If output is ready (synchronous response), extract URL
	if minimaxResp.Output != nil {
		if audioURL, ok := minimaxResp.Output.(string); ok {
			result.AudioURL = audioURL
		}
	}

	m.logger.Info("Music generation submitted",
		zap.String("prediction_id", result.PredictionID),
		zap.String("status", result.Status),
	)

	return result, nil
}

// GetStatus checks the status of a music generation prediction
func (m *MinimaxAdapter) GetStatus(ctx context.Context, predictionID string) (*MusicGenerationResult, error) {
	url := fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", predictionID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+m.apiToken)

	resp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var minimaxResp MinimaxResponse
	if err := json.Unmarshal(body, &minimaxResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := &MusicGenerationResult{
		PredictionID: minimaxResp.ID,
		Status:       minimaxResp.Status,
		Error:        minimaxResp.Error,
	}

	// Extract audio URL if completed
	if minimaxResp.Status == "succeeded" && minimaxResp.Output != nil {
		if audioURL, ok := minimaxResp.Output.(string); ok {
			result.AudioURL = audioURL
		}
	}

	return result, nil
}

// generateMusicPrompt creates a 10-300 character music prompt from video context
func (m *MinimaxAdapter) generateMusicPrompt(videoPrompt, mood, style string) string {
	// Extract key visual elements from video prompt (first few words)
	keywords := extractKeywords(videoPrompt, 3)

	// Build music prompt: style + mood + context
	prompt := fmt.Sprintf("%s %s music", capitalizeFirst(style), mood)

	if keywords != "" {
		prompt = fmt.Sprintf("%s %s %s background music",
			capitalizeFirst(style), mood, keywords)
	}

	// Ensure 10-300 char constraint
	if len(prompt) < 10 {
		prompt = fmt.Sprintf("%s background music for video", prompt)
	}
	if len(prompt) > 300 {
		prompt = prompt[:297] + "..."
	}

	return prompt
}

// generateLyrics creates instrumental structure based on duration
func (m *MinimaxAdapter) generateLyrics(duration int) string {
	// For instrumental music, use structural markers
	if duration <= 15 {
		return "[intro]\n[verse]\n[outro]"
	} else if duration <= 30 {
		return "[intro]\n[verse]\n[chorus]\n[outro]"
	} else if duration <= 60 {
		return "[intro]\n[verse]\n[chorus]\n[verse]\n[chorus]\n[outro]"
	} else {
		// Longer videos
		return "[intro]\n[verse]\n[chorus]\n[bridge]\n[verse]\n[chorus]\n[outro]"
	}
}

// extractKeywords extracts first N meaningful words from prompt
func extractKeywords(prompt string, maxWords int) string {
	// Remove common stop words and take first few meaningful words
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "of": true, "with": true,
	}

	words := strings.Fields(strings.ToLower(prompt))
	keywords := []string{}

	for _, word := range words {
		// Clean word
		word = strings.Trim(word, ".,!?;:")
		if len(word) < 2 || stopWords[word] {
			continue
		}
		keywords = append(keywords, word)
		if len(keywords) >= maxWords {
			break
		}
	}

	return strings.Join(keywords, " ")
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
