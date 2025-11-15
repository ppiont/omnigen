package adapters

import (
	"context"
)

// VideoGenerationRequest represents a video generation request
type VideoGenerationRequest struct {
	Prompt         string
	Duration       int    // in seconds (5 or 10 for Kling)
	AspectRatio    string // "16:9", "9:16", "1:1"
	Style          string // optional style modifiers
	StartImageURL  string // optional: URL to start image (first frame)
	NegativePrompt string // optional: things to avoid in the video
}

// VideoGenerationResult represents the result of a video generation
type VideoGenerationResult struct {
	VideoURL    string
	PredictionID string // ID from the model provider for tracking
	Status      string // "processing", "completed", "failed"
	Error       string // error message if failed
}

// VideoGeneratorAdapter is the interface for video generation models
type VideoGeneratorAdapter interface {
	// GenerateVideo submits a video generation request and returns immediately
	GenerateVideo(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResult, error)

	// GetStatus checks the status of a video generation job
	GetStatus(ctx context.Context, predictionID string) (*VideoGenerationResult, error)

	// GetModelName returns the name of the model
	GetModelName() string

	// GetCostPerSecond returns the approximate cost per second of video
	GetCostPerSecond() float64
}
