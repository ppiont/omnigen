package handlers

import "time"

// Video generation constants
const (
	// VideoGenerationTimeout is the maximum time for entire video generation pipeline
	VideoGenerationTimeout = 15 * time.Minute

	// VideoGenerationMaxAttempts is maximum polling attempts for video generation
	// Veo 3.1 can take 15-20 minutes, especially with images
	// 240 attempts × 5s = 20 minutes
	VideoGenerationMaxAttempts = 240

	// AudioGenerationMaxAttempts is maximum polling attempts for Minimax audio generation (5 minutes @ 5s intervals)
	AudioGenerationMaxAttempts = 60

	// PollInterval is the interval between status polls for external APIs
	PollInterval = 5 * time.Second

	// EstimatedCompletionSeconds is the estimated time for full video generation
	EstimatedCompletionSeconds = 300 // ~5 minutes

	// MaxConcurrentGenerations is the maximum number of concurrent video generations
	MaxConcurrentGenerations = 10
)
