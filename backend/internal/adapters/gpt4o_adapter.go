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

	// Enhanced prompt options (optional)
	EnhancedOptions *prompts.EnhancedPromptOptions

	// Style reference image - will be analyzed and converted to text description
	StyleReferenceImage string
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

	// Analyze style reference image if provided
	var styleDescription string
	if req.StyleReferenceImage != "" {
		var err error
		styleDescription, err = g.AnalyzeStyleReference(ctx, req.StyleReferenceImage)
		if err != nil {
			g.logger.Warn("Failed to analyze style reference image, continuing without it",
				zap.Error(err),
			)
			// Continue without style description rather than failing completely
			styleDescription = ""
		}
	}

	// Build user prompt from request
	userPrompt := buildUserPrompt(req)

	g.logger.Debug("Generated user prompt",
		zap.String("prompt", userPrompt),
	)

	// Build enhanced system prompt if options provided
	systemPrompt := prompts.AdScriptSystemPrompt + "\n\n" + prompts.AdScriptFewShotExamples
	if req.EnhancedOptions != nil {
		systemPrompt = prompts.BuildEnhancedSystemPrompt(systemPrompt, req.EnhancedOptions)
		g.logger.Info("Using enhanced system prompt",
			zap.String("style", req.EnhancedOptions.Style),
			zap.String("tone", req.EnhancedOptions.Tone),
			zap.String("platform", req.EnhancedOptions.Platform),
			zap.Bool("pro_cinematography", req.EnhancedOptions.ProCinematography),
		)
	}

	// Determine temperature based on creative boost
	temperature := 0.7 // Default: creative but not random
	if req.EnhancedOptions != nil && req.EnhancedOptions.CreativeBoost {
		temperature = 0.9 // Boosted creativity
		g.logger.Info("Using creative boost", zap.Float64("temperature", temperature))
	}

	// Build Replicate API request
	gpt4oReq := GPT4oRequest{
		Version: g.modelVersion,
		Input: map[string]interface{}{
			"messages": []map[string]string{
				{
					"role":    "system",
					"content": systemPrompt,
				},
				{
					"role":    "user",
					"content": userPrompt,
				},
			},
			"temperature":           temperature,
			"max_completion_tokens": 8192, // Sufficient for complex 60s scripts with multiple scenes
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
	// Don't use Prefer: wait to avoid 60-second timeout - we'll poll instead
	// httpReq.Header.Set("Prefer", "wait")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var gpt4oResp GPT4oResponse
	if err := json.Unmarshal(body, &gpt4oResp); err != nil {
		g.logger.Error("Failed to unmarshal initial response",
			zap.Error(err),
			zap.String("body_preview", string(body)[:min(1000, len(body))]),
		)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	g.logger.Info("Initial GPT-4o response",
		zap.String("prediction_id", gpt4oResp.ID),
		zap.String("status", gpt4oResp.Status),
		zap.Int("initial_output_chunks", len(gpt4oResp.Output)),
	)

	// If output not ready, poll for completion
	if gpt4oResp.Status != "succeeded" {
		g.logger.Info("Waiting for GPT-4o completion",
			zap.String("prediction_id", gpt4oResp.ID),
			zap.String("status", gpt4oResp.Status),
		)

		// Poll for completion (max 5 minutes for script generation)
		// Longer timeout because script generation can be complex with vision analysis
		maxAttempts := 60 // 60 * 5s = 5 minutes
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
				g.logger.Warn("Failed to poll status, retrying",
					zap.Error(err),
					zap.Int("attempt", attempt+1),
					zap.Int("max_attempts", maxAttempts),
				)
				continue
			}

			g.logger.Info("Poll status check",
				zap.String("status", pollResp.Status),
				zap.Int("output_length", len(pollResp.Output)),
				zap.Int("attempt", attempt+1),
			)

			if pollResp.Status == "succeeded" && len(pollResp.Output) > 0 {
				gpt4oResp = *pollResp
				break
			}

			if pollResp.Status == "failed" || pollResp.Status == "canceled" {
				return nil, fmt.Errorf("GPT-4o generation failed: %s", pollResp.Error)
			}

			// Update response for next iteration even if not succeeded yet
			gpt4oResp = *pollResp
		}
	}

	// Extract and parse JSON response
	if len(gpt4oResp.Output) == 0 {
		return nil, fmt.Errorf("no output from GPT-4o after polling (final status: %s)", gpt4oResp.Status)
	}

	// Check if we timed out without succeeding
	if gpt4oResp.Status != "succeeded" {
		return nil, fmt.Errorf("GPT-4o generation did not complete in time (status: %s)", gpt4oResp.Status)
	}

	// Combine output array into single string
	var scriptJSON string
	for _, part := range gpt4oResp.Output {
		scriptJSON += part
	}

	g.logger.Info("Received GPT-4o output",
		zap.Int("output_length", len(scriptJSON)),
		zap.Int("output_chunks", len(gpt4oResp.Output)),
		zap.String("output_preview", scriptJSON[:min(500, len(scriptJSON))]),
		zap.String("output_end", scriptJSON[max(0, len(scriptJSON)-200):]),
	)

	// Parse JSON into Script struct
	script, err := g.parseScriptJSON(scriptJSON, styleDescription)
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

	// Use a separate client with shorter timeout for polling requests
	pollClient := &http.Client{
		Timeout: 30 * time.Second, // Each poll request should complete quickly
	}

	resp, err := pollClient.Do(httpReq)
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

// AnalyzeStyleReference uses GPT-4o Vision to analyze a reference image and extract style description
func (g *GPT4oAdapter) AnalyzeStyleReference(ctx context.Context, imageURL string) (string, error) {
	g.logger.Info("Analyzing style reference image with GPT-4o Vision",
		zap.String("image_url", imageURL[:min(100, len(imageURL))]),
	)

	// Build vision request with image
	gpt4oReq := GPT4oRequest{
		Version: g.modelVersion,
		Input: map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": `Analyze this image and describe its visual style in detail for video generation. Focus on:

1. **Color Palette**: Dominant colors, color grading, saturation level
2. **Lighting**: Lighting style (natural, dramatic, soft, hard), shadows, highlights
3. **Mood & Atmosphere**: Overall feeling, emotional tone
4. **Composition**: Framing style, visual balance, focal points
5. **Texture & Detail**: Surface qualities, level of detail, sharpness
6. **Cinematography**: Camera feel (static, dynamic), depth of field, perspective

Provide a concise 2-3 sentence description that captures the essence of this visual style, suitable for adding to video generation prompts.`,
						},
						{
							"type": "image_url",
							"image_url": map[string]string{
								"url": imageURL,
							},
						},
					},
				},
			},
			"temperature":           0.3, // Lower temperature for consistent style analysis
			"max_completion_tokens": 500,
		},
	}

	payload, err := json.Marshal(gpt4oReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Submit prediction to Replicate
	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.replicate.com/v1/predictions",
		bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+g.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	// Don't use Prefer: wait to avoid 60-second timeout - we'll poll instead
	// httpReq.Header.Set("Prefer", "wait")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var gpt4oResp GPT4oResponse
	if err := json.Unmarshal(body, &gpt4oResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// If output not ready, poll for completion
	if gpt4oResp.Status != "succeeded" {
		g.logger.Info("Waiting for GPT-4o Vision analysis",
			zap.String("prediction_id", gpt4oResp.ID),
		)

		// Poll for completion (max 2 minutes for vision analysis)
		maxAttempts := 24 // 24 * 5s = 2 minutes
		pollInterval := 5 * time.Second

		for attempt := 0; attempt < maxAttempts; attempt++ {
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("context cancelled: %w", ctx.Err())
			default:
			}

			time.Sleep(pollInterval)

			pollResp, err := g.pollStatus(ctx, gpt4oResp.ID)
			if err != nil {
				g.logger.Warn("Failed to poll vision status, retrying",
					zap.Error(err),
					zap.Int("attempt", attempt+1),
					zap.Int("max_attempts", maxAttempts),
				)
				continue
			}

			g.logger.Info("Vision poll status check",
				zap.String("status", pollResp.Status),
				zap.Int("output_length", len(pollResp.Output)),
				zap.Int("attempt", attempt+1),
			)

			if pollResp.Status == "succeeded" && len(pollResp.Output) > 0 {
				gpt4oResp = *pollResp
				break
			}

			if pollResp.Status == "failed" || pollResp.Status == "canceled" {
				return "", fmt.Errorf("Vision analysis failed: %s", pollResp.Error)
			}

			// Update response for next iteration
			gpt4oResp = *pollResp
		}
	}

	// Extract style description from output
	if len(gpt4oResp.Output) == 0 {
		return "", fmt.Errorf("no output from GPT-4o Vision after polling (final status: %s)", gpt4oResp.Status)
	}

	// Check if we timed out without succeeding
	if gpt4oResp.Status != "succeeded" {
		return "", fmt.Errorf("GPT-4o Vision analysis did not complete in time (status: %s)", gpt4oResp.Status)
	}

	// Concatenate all output chunks (GPT-4o streams response)
	var styleDescription string
	for _, chunk := range gpt4oResp.Output {
		styleDescription += chunk
	}

	g.logger.Info("Style analysis complete",
		zap.String("style_description", styleDescription[:min(200, len(styleDescription))]),
	)

	return styleDescription, nil
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
func (g *GPT4oAdapter) parseScriptJSON(scriptJSON string, styleDescription string) (*domain.Script, error) {
	// Try to extract JSON if wrapped in markdown code blocks
	cleaned := extractJSON(scriptJSON)

	var script domain.Script
	if err := json.Unmarshal([]byte(cleaned), &script); err != nil {
		return nil, fmt.Errorf("failed to unmarshal script: %w (JSON: %s)", err, cleaned[:min(200, len(cleaned))])
	}

	// Add style description to script and append to each scene's generation_prompt
	if styleDescription != "" {
		script.StyleDescription = styleDescription
		g.logger.Info("Appending style description to scene prompts",
			zap.String("style_description", styleDescription[:min(150, len(styleDescription))]),
			zap.Int("num_scenes", len(script.Scenes)),
		)

		for i := range script.Scenes {
			// Append style description to each scene's generation_prompt
			script.Scenes[i].GenerationPrompt = script.Scenes[i].GenerationPrompt + ". Style: " + styleDescription
		}
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

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
