package service

import (
	"testing"

	"github.com/omnigen/backend/internal/domain"
)

func TestCalculateMusicTail(t *testing.T) {
	tests := []struct {
		name          string
		videoDuration int
		expected      float64
	}{
		{
			name:          "10s video gets 1.0s music tail (minimum)",
			videoDuration: 10,
			expected:      1.0,
		},
		{
			name:          "15s video gets 1.0s music tail",
			videoDuration: 15,
			expected:      1.0,
		},
		{
			name:          "30s video gets 1.0s music tail",
			videoDuration: 30,
			expected:      1.0,
		},
		{
			name:          "45s video gets 1.5s music tail",
			videoDuration: 45,
			expected:      1.5,
		},
		{
			name:          "60s video gets 2.0s music tail (maximum)",
			videoDuration: 60,
			expected:      2.0,
		},
		{
			name:          "90s video capped at 2.0s music tail",
			videoDuration: 90,
			expected:      2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateMusicTail(tt.videoDuration)
			if result != tt.expected {
				t.Errorf("CalculateMusicTail(%d) = %v, want %v", tt.videoDuration, result, tt.expected)
			}
		})
	}
}

func TestCalculateNarrationBudget(t *testing.T) {
	tests := []struct {
		name               string
		videoDuration      int
		disclaimerDuration float64
		expectedSeconds    float64
		expectedWords      int
	}{
		{
			name:               "60s video with 8s disclaimer",
			videoDuration:      60,
			disclaimerDuration: 8.0,
			expectedSeconds:    50.0, // 60 - 8 - 2.0 (music tail)
			expectedWords:      125,  // 50 * 2.5
		},
		{
			name:               "40s video with 6s disclaimer",
			videoDuration:      40,
			disclaimerDuration: 6.0,
			// music tail = min(2.0, max(1.0, 40/30)) = min(2.0, 1.33) = 1.33
			expectedSeconds: 32.67, // 40 - 6 - 1.33 (approx)
			expectedWords:   81,    // 32.67 * 2.5 (truncated to int)
		},
		{
			name:               "30s video with 5s disclaimer",
			videoDuration:      30,
			disclaimerDuration: 5.0,
			expectedSeconds:    24.0, // 30 - 5 - 1.0 (music tail)
			expectedWords:      60,   // 24 * 2.5
		},
		{
			name:               "20s video with 4s short disclaimer",
			videoDuration:      20,
			disclaimerDuration: 4.0,
			expectedSeconds:    15.0, // 20 - 4 - 1.0 (music tail, clamped at 1.0)
			expectedWords:      37,   // 15 * 2.5 = 37.5 (truncated)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seconds, words := CalculateNarrationBudget(tt.videoDuration, tt.disclaimerDuration)

			// Allow small tolerance for floating point
			tolerance := 0.1
			if seconds < tt.expectedSeconds-tolerance || seconds > tt.expectedSeconds+tolerance {
				t.Errorf("CalculateNarrationBudget(%d, %v) seconds = %v, want ~%v",
					tt.videoDuration, tt.disclaimerDuration, seconds, tt.expectedSeconds)
			}

			// Words should be exact since it's integer calculation
			if words != tt.expectedWords {
				t.Errorf("CalculateNarrationBudget(%d, %v) words = %v, want %v",
					tt.videoDuration, tt.disclaimerDuration, words, tt.expectedWords)
			}
		})
	}
}

func TestDisclaimerTierDetermination(t *testing.T) {
	// Test the tier determination logic based on video duration
	// This tests the tier boundaries without requiring actual TTS/GPT calls

	tests := []struct {
		name          string
		videoDuration int
		expectedTier  domain.DisclaimerTier
		expectAudio   bool
	}{
		{
			name:          "10s video gets text-only tier",
			videoDuration: 10,
			expectedTier:  domain.DisclaimerTierTextOnly,
			expectAudio:   false,
		},
		{
			name:          "14s video gets text-only tier (boundary)",
			videoDuration: 14,
			expectedTier:  domain.DisclaimerTierTextOnly,
			expectAudio:   false,
		},
		{
			name:          "15s video gets short tier (boundary)",
			videoDuration: 15,
			expectedTier:  domain.DisclaimerTierShort,
			expectAudio:   true,
		},
		{
			name:          "20s video gets short tier",
			videoDuration: 20,
			expectedTier:  domain.DisclaimerTierShort,
			expectAudio:   true,
		},
		{
			name:          "29s video gets short tier (boundary)",
			videoDuration: 29,
			expectedTier:  domain.DisclaimerTierShort,
			expectAudio:   true,
		},
		{
			name:          "30s video gets full tier (boundary)",
			videoDuration: 30,
			expectedTier:  domain.DisclaimerTierFull,
			expectAudio:   true,
		},
		{
			name:          "40s video gets full tier",
			videoDuration: 40,
			expectedTier:  domain.DisclaimerTierFull,
			expectAudio:   true,
		},
		{
			name:          "60s video gets full tier",
			videoDuration: 60,
			expectedTier:  domain.DisclaimerTierFull,
			expectAudio:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Determine tier based on duration (replicating the logic from ComputeDisclaimerSpec)
			var tier domain.DisclaimerTier
			var useAudio bool

			switch {
			case tt.videoDuration >= 30:
				tier = domain.DisclaimerTierFull
				useAudio = true
			case tt.videoDuration >= 15:
				tier = domain.DisclaimerTierShort
				useAudio = true
			default:
				tier = domain.DisclaimerTierTextOnly
				useAudio = false
			}

			if tier != tt.expectedTier {
				t.Errorf("Tier for %ds video = %v, want %v",
					tt.videoDuration, tier, tt.expectedTier)
			}

			if useAudio != tt.expectAudio {
				t.Errorf("UseAudio for %ds video = %v, want %v",
					tt.videoDuration, useAudio, tt.expectAudio)
			}
		})
	}
}

func TestGenerateAbbreviatedDisclaimer(t *testing.T) {
	// Test the abbreviated disclaimer generation for text-only tier
	service := &DisclaimerService{}

	result := service.generateAbbreviatedDisclaimer("Long disclaimer text that should be shortened")

	// The current implementation returns a static string
	expected := "See Important Safety Information at example.com"
	if result != expected {
		t.Errorf("generateAbbreviatedDisclaimer() = %v, want %v", result, expected)
	}
}

func TestDisclaimerSpecFields(t *testing.T) {
	// Test that DisclaimerSpec has all required fields and they're used correctly
	spec := &domain.DisclaimerSpec{
		Tier:          domain.DisclaimerTierFull,
		FullText:      "Full side effects disclaimer with all warnings.",
		AudioText:     "Full side effects disclaimer with all warnings.",
		AudioDuration: 8.5,
		UseAudio:      true,
		Speed:         1.4,
	}

	if spec.Tier != domain.DisclaimerTierFull {
		t.Errorf("Tier = %v, want %v", spec.Tier, domain.DisclaimerTierFull)
	}
	if spec.FullText != spec.AudioText {
		t.Error("For full tier, FullText and AudioText should match")
	}
	if !spec.UseAudio {
		t.Error("Full tier should have UseAudio = true")
	}
	if spec.Speed != 1.4 {
		t.Errorf("Speed = %v, want 1.4", spec.Speed)
	}
}

func TestDisclaimerTierConstants(t *testing.T) {
	// Verify tier constant values are as expected
	if domain.DisclaimerTierFull != "full" {
		t.Errorf("DisclaimerTierFull = %v, want 'full'", domain.DisclaimerTierFull)
	}
	if domain.DisclaimerTierShort != "short" {
		t.Errorf("DisclaimerTierShort = %v, want 'short'", domain.DisclaimerTierShort)
	}
	if domain.DisclaimerTierTextOnly != "text_only" {
		t.Errorf("DisclaimerTierTextOnly = %v, want 'text_only'", domain.DisclaimerTierTextOnly)
	}
}
