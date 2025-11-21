package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/omnigen/backend/internal/adapters"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/prompts"
)

// ParserService handles ad script generation
type ParserService struct {
	gpt4o  *adapters.GPT4oAdapter
	logger *zap.Logger
}

// ParseRequest represents user input for script generation - SIMPLE interface
type ParseRequest struct {
	UserID      string `json:"user_id"`
	Prompt      string `json:"prompt"`                // Free-form user input with ALL context
	Duration    int    `json:"duration"`              // 10-60 seconds (must be multiple of 10)
	AspectRatio string `json:"aspect_ratio"`          // "16:9", "9:16", or "1:1"
	StartImage  string `json:"start_image,omitempty"` // Optional starting image URL (first scene only)

	// Style reference image - analyzed and converted to text description for ALL scenes
	StyleReferenceImage string `json:"style_reference_image,omitempty"`

	// Pharmaceutical ad configuration
	Voice       string `json:"voice,omitempty"`       // "male" or "female" for narrator
	SideEffects string `json:"side_effects,omitempty"` // User-provided side effects disclosure text

	// Enhanced prompt options (Phase 1 - all optional)
	Style             string
	Tone              string
	Tempo             string
	Platform          string
	Audience          string
	Goal              string
	CallToAction      string
	ProCinematography bool
	CreativeBoost     bool
}

// NewParserService creates a new script parser service
func NewParserService(
	gpt4o *adapters.GPT4oAdapter,
	logger *zap.Logger,
) *ParserService {
	return &ParserService{
		gpt4o:  gpt4o,
		logger: logger,
	}
}

// GenerateScript creates a new ad script using GPT-4o
func (s *ParserService) GenerateScript(ctx context.Context, req ParseRequest) (*domain.Script, error) {
	s.logger.Info("Generating script with GPT-4o",
		zap.String("user_id", req.UserID),
		zap.Int("duration", req.Duration),
		zap.String("prompt", req.Prompt))

	// Validate request
	if err := s.validateParseRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Build enhanced prompt options if any are provided
	var enhancedOptions *prompts.EnhancedPromptOptions
	if req.Style != "" || req.Tone != "" || req.Tempo != "" || req.Platform != "" ||
		req.Audience != "" || req.Goal != "" || req.CallToAction != "" || req.ProCinematography || req.CreativeBoost {
		enhancedOptions = &prompts.EnhancedPromptOptions{
			Style:             req.Style,
			Tone:              req.Tone,
			Tempo:             req.Tempo,
			Platform:          req.Platform,
			Audience:          req.Audience,
			Goal:              req.Goal,
			CallToAction:      req.CallToAction,
			ProCinematography: req.ProCinematography,
			CreativeBoost:     req.CreativeBoost,
		}
	}

	// Call GPT-4o adapter - GPT-4o will extract product info from prompt
	gpt4oReq := &adapters.ScriptGenerationRequest{
		Prompt:              req.Prompt,
		Duration:            req.Duration,
		AspectRatio:         req.AspectRatio,
		StartImage:          req.StartImage,
		StyleReferenceImage: req.StyleReferenceImage,
		Voice:               req.Voice,
		SideEffects:         req.SideEffects,
		EnhancedOptions:     enhancedOptions,
	}

	script, err := s.gpt4o.GenerateScript(ctx, gpt4oReq)
	if err != nil {
		return nil, fmt.Errorf("GPT-4o generation failed: %w", err)
	}

	// Script will be embedded in Job - no separate persistence needed
	s.logger.Info("Script generated successfully",
		zap.Int("num_scenes", len(script.Scenes)),
		zap.String("title", script.Title))

	return script, nil
}

// validateParseRequest validates the parse request
func (s *ParserService) validateParseRequest(req ParseRequest) error {
	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	if req.Duration < 10 || req.Duration > 60 {
		return fmt.Errorf("duration must be between 10 and 60 seconds")
	}

	if req.Duration%10 != 0 {
		return fmt.Errorf("duration must be a multiple of 10 seconds (Kling constraint)")
	}

	if req.AspectRatio != "16:9" && req.AspectRatio != "9:16" && req.AspectRatio != "1:1" {
		return fmt.Errorf("aspect_ratio must be one of: 16:9, 9:16, 1:1")
	}

	return nil
}
