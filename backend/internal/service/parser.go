package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/omnigen/backend/internal/domain"
	"go.uber.org/zap"
)

// PromptParser parses user prompts to extract video generation parameters
type PromptParser struct {
	claudeAPIKey string
	httpClient   *http.Client
	logger       *zap.Logger
}

// NewPromptParser creates a new prompt parser
func NewPromptParser(logger *zap.Logger) *PromptParser {
	return &PromptParser{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// ParsePrompt parses a user prompt and extracts key parameters
func (p *PromptParser) ParsePrompt(ctx context.Context, req *domain.GenerateRequest) (*domain.ParsedPrompt, error) {
	// For MVP, use regex-based parsing with fallback to Claude API
	parsed := p.parseWithRegex(req)

	p.logger.Info("Prompt parsed successfully",
		zap.String("prompt", req.Prompt),
		zap.String("product_type", parsed.ProductType),
		zap.Strings("visual_style", parsed.VisualStyle),
	)

	return parsed, nil
}

// parseWithRegex uses simple regex and keyword matching for MVP
func (p *PromptParser) parseWithRegex(req *domain.GenerateRequest) *domain.ParsedPrompt {
	prompt := strings.ToLower(req.Prompt)

	parsed := &domain.ParsedPrompt{
		Duration:     req.Duration,
		AspectRatio:  req.AspectRatio,
		VisualStyle:  []string{},
		ColorPalette: []string{},
		TextOverlays: []string{},
	}

	// If aspect ratio not specified, default to 16:9
	if parsed.AspectRatio == "" {
		parsed.AspectRatio = domain.AspectRatio16x9
	}

	// Extract product type
	productTypes := map[string]string{
		"watch":    "luxury watch",
		"perfume":  "perfume",
		"car":      "car",
		"phone":    "smartphone",
		"laptop":   "laptop",
		"shoes":    "shoes",
		"clothing": "clothing",
		"jewelry":  "jewelry",
		"food":     "food product",
		"drink":    "beverage",
	}

	for keyword, productType := range productTypes {
		if strings.Contains(prompt, keyword) {
			parsed.ProductType = productType
			break
		}
	}

	// If no specific product found, use generic
	if parsed.ProductType == "" {
		parsed.ProductType = "product"
	}

	// Extract visual styles
	styleKeywords := []string{
		"luxury", "minimal", "elegant", "modern", "vintage",
		"futuristic", "retro", "professional", "playful",
		"sophisticated", "bold", "clean", "artistic",
	}

	for _, style := range styleKeywords {
		if strings.Contains(prompt, style) {
			parsed.VisualStyle = append(parsed.VisualStyle, style)
		}
	}

	// Add style from request if provided
	if req.Style != "" {
		styles := strings.Split(req.Style, ",")
		for _, s := range styles {
			s = strings.TrimSpace(s)
			if s != "" && !contains(parsed.VisualStyle, s) {
				parsed.VisualStyle = append(parsed.VisualStyle, s)
			}
		}
	}

	// Default to modern if no style found
	if len(parsed.VisualStyle) == 0 {
		parsed.VisualStyle = []string{"modern", "professional"}
	}

	// Extract color palette
	colorKeywords := []string{
		"gold", "silver", "black", "white", "blue", "red",
		"green", "purple", "pink", "orange", "brown",
		"rose gold", "metallic",
	}

	for _, color := range colorKeywords {
		if strings.Contains(prompt, color) {
			parsed.ColorPalette = append(parsed.ColorPalette, color)
		}
	}

	// Extract text overlays (CTA, product name)
	// Look for quoted text
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindAllStringSubmatch(req.Prompt, -1)
	for _, match := range matches {
		if len(match) > 1 {
			parsed.TextOverlays = append(parsed.TextOverlays, match[1])
		}
	}

	return parsed
}

// parseWithClaude uses Claude API for advanced prompt parsing (future enhancement)
func (p *PromptParser) parseWithClaude(ctx context.Context, prompt string) (*domain.ParsedPrompt, error) {
	// This is a placeholder for Claude API integration
	// For MVP, we're using regex-based parsing

	type ClaudeRequest struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		Messages  []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	systemPrompt := `You are a video ad creative parser. Extract the following from the user's prompt:
- product_type: what product is being advertised
- visual_style: array of style keywords (luxury, minimal, modern, etc.)
- color_palette: array of colors mentioned
- text_overlays: any text that should appear in the video

Respond in JSON format only.`

	reqBody := ClaudeRequest{
		Model:     "claude-3-haiku-20240307",
		MaxTokens: 500,
	}
	reqBody.Messages = []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{
		{Role: "user", Content: systemPrompt + "\n\nPrompt: " + prompt},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.claudeAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call claude API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("claude API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Claude response and extract the parsed prompt
	// This is simplified - actual implementation would parse Claude's response format
	var parsed domain.ParsedPrompt
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &parsed, nil
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
