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

	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/prompts"
)

// GPT4oAdapter implements script generation via OpenAI GPT-4o on Replicate
type GPT4oAdapter struct {
	apiToken     string
	httpClient   *http.Client
	logger       *zap.Logger
	modelVersion string
}

// NewGPT4oAdapter creates a new GPT-4o adapter
func NewGPT4oAdapter(apiToken string, logger *zap.Logger) *GPT4oAdapter {
	return &GPT4oAdapter{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // GPT-4o can take a while for complex scripts
		},
		logger:       logger,
		modelVersion: "openai/gpt-4o:ad45308bffd6defaaa05dff12658b454a3a8dcfd7cc1440420a74d87a48caa9e",
	}
}

// ScriptGenerationRequest represents the input for script generation - SIMPLE interface
type ScriptGenerationRequest struct {
	Prompt      string // Free-form prompt with ALL context (product, audience, vibe, etc.)
	Duration    int    // Total duration in seconds
	AspectRatio string // "16:9", "9:16", or "1:1"
	StartImage  string // Optional starting image URL for first scene
}

// GPT4oRequest matches the Replicate OpenAI GPT-4o API schema
type GPT4oRequest struct {
	Version string                 `json:"version"`
	Input   map[string]interface{} `json:"input"`
}

// GPT4oResponse represents the Replicate API response
type GPT4oResponse struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Output []string `json:"output,omitempty"` // Array of strings (streaming)
	Error  string   `json:"error,omitempty"`
}

// GenerateScript generates a structured ad script using GPT-4o
func (g *GPT4oAdapter) GenerateScript(ctx context.Context, req *ScriptGenerationRequest) (*domain.Script, error) {
	g.logger.Info("Generating script with GPT-4o",
		zap.String("prompt", req.Prompt[:min(100, len(req.Prompt))]),
		zap.Int("duration", req.Duration),
	)

	// Build user prompt from request
	userPrompt := buildUserPrompt(req)

	g.logger.Debug("Generated user prompt",
		zap.String("prompt", userPrompt),
	)

	// Build Replicate API request
	gpt4oReq := GPT4oRequest{
		Version: g.modelVersion,
		Input: map[string]interface{}{
			"messages": []map[string]string{
				{
					"role":    "system",
					"content": prompts.AdScriptSystemPrompt + "\n\n" + prompts.AdScriptFewShotExamples,
				},
				{
					"role":    "user",
					"content": userPrompt,
				},
			},
			"temperature":           0.7,   // Creative but not random
			"max_completion_tokens": 16384, // Increased for complex scripts with many scenes
			"top_p":                 0.9,
		},
	}

	payload, err := json.Marshal(gpt4oReq)
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

	httpReq.Header.Set("Authorization", "Bearer "+g.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Prefer", "wait") // Wait for completion if possible

	resp, err := g.httpClient.Do(httpReq)
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

	var gpt4oResp GPT4oResponse
	if err := json.Unmarshal(body, &gpt4oResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// If output not ready, poll for completion
	if gpt4oResp.Status != "succeeded" && gpt4oResp.Output == nil {
		g.logger.Info("Waiting for GPT-4o completion",
			zap.String("prediction_id", gpt4oResp.ID),
			zap.String("status", gpt4oResp.Status),
		)

		// Poll for completion (max 2 minutes)
		maxAttempts := 24 // 24 * 5s = 2 minutes
		pollInterval := 5 * time.Second

		for attempt := 0; attempt < maxAttempts; attempt++ {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled while polling: %w", ctx.Err())
			default:
			}

			time.Sleep(pollInterval)

			pollResp, err := g.pollStatus(ctx, gpt4oResp.ID)
			if err != nil {
				g.logger.Warn("Failed to poll status, retrying", zap.Error(err))
				continue
			}

			if pollResp.Status == "succeeded" && len(pollResp.Output) > 0 {
				gpt4oResp = *pollResp
				break
			}

			if pollResp.Status == "failed" || pollResp.Status == "canceled" {
				return nil, fmt.Errorf("GPT-4o generation failed: %s", pollResp.Error)
			}
		}
	}

	// Extract and parse JSON response
	if len(gpt4oResp.Output) == 0 {
		return nil, fmt.Errorf("no output from GPT-4o")
	}

	// Combine output array into single string
	var scriptJSON string
	for _, part := range gpt4oResp.Output {
		scriptJSON += part
	}

	g.logger.Debug("Received GPT-4o output",
		zap.String("output", scriptJSON[:min(500, len(scriptJSON))]),
	)

	// Parse JSON into Script struct
	script, err := parseScriptJSON(scriptJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse script JSON: %w", err)
	}

	// Validate script
	if err := validateScript(script, req.Duration); err != nil {
		return nil, fmt.Errorf("script validation failed: %w", err)
	}

	g.logger.Info("Script generated successfully",
		zap.String("title", script.Title),
		zap.Int("num_scenes", len(script.Scenes)),
		zap.Int("total_duration", script.TotalDuration),
	)

	return script, nil
}

// pollStatus checks the status of a GPT-4o prediction
func (g *GPT4oAdapter) pollStatus(ctx context.Context, predictionID string) (*GPT4oResponse, error) {
	url := fmt.Sprintf("https://api.replicate.com/v1/predictions/%s", predictionID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+g.apiToken)

	resp, err := g.httpClient.Do(httpReq)
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

	var gpt4oResp GPT4oResponse
	if err := json.Unmarshal(body, &gpt4oResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &gpt4oResp, nil
}

// buildUserPrompt constructs the user prompt from request parameters
func buildUserPrompt(req *ScriptGenerationRequest) string {
	prompt := fmt.Sprintf(`Create a %d-second advertisement video script based on this creative direction:

%s

**Aspect Ratio:** %s`,
		req.Duration,
		req.Prompt,
		req.AspectRatio,
	)

	if req.StartImage != "" {
		prompt += fmt.Sprintf("\n**Starting Image:** %s (use as reference for first scene)", req.StartImage)
	}

	prompt += `

**Instructions:**
- Extract product details, target audience, and brand vibe from the creative direction above
- Generate varied scenes with different shot types, angles, and lighting (NO repetitive clips!)
- Each scene must have a unique, detailed generation_prompt optimized for Kling AI
- Derive appropriate music_mood and music_style based on the content
- Return ONLY valid JSON matching the Script schema, no markdown or explanations`

	return prompt
}

// parseScriptJSON parses the GPT-4o JSON output into a Script struct
func parseScriptJSON(scriptJSON string) (*domain.Script, error) {
	// Try to extract JSON if wrapped in markdown code blocks
	cleaned := extractJSON(scriptJSON)

	var script domain.Script
	if err := json.Unmarshal([]byte(cleaned), &script); err != nil {
		return nil, fmt.Errorf("failed to unmarshal script: %w (JSON: %s)", err, cleaned[:min(200, len(cleaned))])
	}

	return &script, nil
}

// extractJSON extracts JSON from markdown code blocks if present
func extractJSON(s string) string {
	// Remove markdown code blocks if present
	if len(s) > 7 && s[:3] == "```" {
		// Find first newline after ```
		start := 3
		for start < len(s) && s[start] != '\n' {
			start++
		}
		start++

		// Find closing ```
		end := len(s)
		if idx := bytes.Index([]byte(s[start:]), []byte("```")); idx != -1 {
			end = start + idx
		}

		return s[start:end]
	}

	return s
}

// validateScript ensures the generated script meets requirements
func validateScript(script *domain.Script, requestedDuration int) error {
	if script.Title == "" {
		return fmt.Errorf("script title is empty")
	}

	if len(script.Scenes) == 0 {
		return fmt.Errorf("script has no scenes")
	}

	// Validate total duration matches request (allow 10% variance)
	if script.TotalDuration < requestedDuration-5 || script.TotalDuration > requestedDuration+5 {
		return fmt.Errorf("total duration %d doesn't match requested %d", script.TotalDuration, requestedDuration)
	}

	// Validate scenes
	var totalDuration float64
	for i, scene := range script.Scenes {
		if scene.SceneNumber != i+1 {
			return fmt.Errorf("scene %d has incorrect scene_number %d", i+1, scene.SceneNumber)
		}

		if scene.GenerationPrompt == "" {
			return fmt.Errorf("scene %d has empty generation_prompt", i+1)
		}

		totalDuration += scene.Duration
	}

	// Validate audio spec
	if script.AudioSpec.MusicMood == "" {
		return fmt.Errorf("audio_spec missing music_mood")
	}

	if script.AudioSpec.MusicStyle == "" {
		return fmt.Errorf("audio_spec missing music_style")
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
