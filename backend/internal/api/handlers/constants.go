package handlers

import "time"

// Video generation constants
const (
	// VideoGenerationTimeout is the maximum time for entire video generation pipeline
	VideoGenerationTimeout = 15 * time.Minute

	// VideoGenerationMaxAttempts is maximum polling attempts for Kling video generation (10 minutes @ 5s intervals)
	VideoGenerationMaxAttempts = 120

	// AudioGenerationMaxAttempts is maximum polling attempts for Minimax audio generation (5 minutes @ 5s intervals)
	AudioGenerationMaxAttempts = 60

	// PollInterval is the interval between status polls for external APIs
	PollInterval = 5 * time.Second

	// EstimatedCompletionSeconds is the estimated time for full video generation
	EstimatedCompletionSeconds = 300 // ~5 minutes

	// MaxConcurrentGenerations is the maximum number of concurrent video generations
	MaxConcurrentGenerations = 10
)
