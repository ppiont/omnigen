package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/omnigen/backend/internal/adapters"
	"github.com/omnigen/backend/internal/domain"
	"go.uber.org/zap"
)

// DisclaimerService handles tiered disclaimer logic and timing calculations.
type DisclaimerService struct {
	ttsAdapter *adapters.OpenAITTSAdapter
	gptAdapter *adapters.GPT4oAdapter
	logger     *zap.Logger
}

// NewDisclaimerService creates a new disclaimer service instance.
func NewDisclaimerService(
	tts *adapters.OpenAITTSAdapter,
	gpt *adapters.GPT4oAdapter,
	logger *zap.Logger,
) *DisclaimerService {
	return &DisclaimerService{
		ttsAdapter: tts,
		gptAdapter: gpt,
		logger:     logger,
	}
}

// ComputeDisclaimerSpec determines the appropriate disclaimer tier and generates audio timing.
func (s *DisclaimerService) ComputeDisclaimerSpec(
	ctx context.Context,
	fullDisclaimerText string,
	videoDuration int,
	voice string,
) (*domain.DisclaimerSpec, error) {
	spec := &domain.DisclaimerSpec{
		FullText: fullDisclaimerText,
	}

	s.logger.Info("Computing disclaimer spec",
		zap.Int("video_duration", videoDuration),
		zap.Int("full_text_length", len(fullDisclaimerText)),
	)

	// Determine tier based on video length
	switch {
	case videoDuration >= 30:
		spec.Tier = domain.DisclaimerTierFull
		spec.AudioText = fullDisclaimerText
		spec.Speed = 1.4
		spec.UseAudio = true

	case videoDuration >= 15:
		spec.Tier = domain.DisclaimerTierShort
		spec.Speed = 1.4
		spec.UseAudio = true
		// Generate short version
		shortText, err := s.generateShortDisclaimer(ctx, fullDisclaimerText)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short disclaimer: %w", err)
		}
		spec.AudioText = shortText

	default: // <15s
		spec.Tier = domain.DisclaimerTierTextOnly
		spec.UseAudio = false
		spec.AudioDuration = 0
		// For text-only, use abbreviated version for overlay
		spec.AudioText = s.generateAbbreviatedDisclaimer(fullDisclaimerText)

		s.logger.Info("Using text-only disclaimer tier",
			zap.String("tier", string(spec.Tier)),
			zap.String("abbreviated_text", spec.AudioText),
		)
		return spec, nil
	}

	// Generate TTS and get duration
	_, duration, err := s.ttsAdapter.GenerateVoiceoverWithDuration(ctx, spec.AudioText, voice, spec.Speed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate disclaimer TTS: %w", err)
	}
	spec.AudioDuration = duration

	s.logger.Info("Disclaimer spec computed",
		zap.String("tier", string(spec.Tier)),
		zap.Float64("audio_duration", spec.AudioDuration),
		zap.Int("audio_text_words", len(strings.Fields(spec.AudioText))),
		zap.Float64("speed", spec.Speed),
	)

	return spec, nil
}

// generateShortDisclaimer creates a 12-18 word version of the full disclaimer using GPT.
func (s *DisclaimerService) generateShortDisclaimer(ctx context.Context, fullText string) (string, error) {
	systemPrompt := "You are a pharmaceutical regulatory copywriter. You create compliant, concise safety disclosures."

	userPrompt := fmt.Sprintf(`Here is the full Important Safety Information:

%s

Generate a shortened audio disclaimer version suitable for 15-20 second ads (12–18 words max).
Keep ALL contraindications and the 3–4 most common/serious side effects.
Start with "Do not take..." or "May cause..." phrasing.
Make it sound natural when spoken fast.
Output only the short text, nothing else.`, fullText)

	s.logger.Info("Generating short disclaimer via GPT",
		zap.Int("full_text_words", len(strings.Fields(fullText))),
	)

	// Call GPT to generate short version
	shortText, err := s.gptAdapter.GenerateText(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	shortText = strings.TrimSpace(shortText)
	wordCount := len(strings.Fields(shortText))

	s.logger.Info("Short disclaimer generated",
		zap.Int("word_count", wordCount),
		zap.String("short_text", shortText),
	)

	return shortText, nil
}

// generateAbbreviatedDisclaimer creates a 6-10 word version for text-only overlay.
func (s *DisclaimerService) generateAbbreviatedDisclaimer(fullText string) string {
	// Simple extraction: "See Important Safety Information"
	// Could be enhanced to extract key warning
	return "See Important Safety Information at example.com"
}

// CalculateMusicTail returns appropriate music tail duration based on video length.
func CalculateMusicTail(videoDuration int) float64 {
	// 1.0s for 15-30s, scaling up to 2.0s for 60s+
	tail := math.Min(2.0, math.Max(1.0, float64(videoDuration)/30.0))
	return tail
}

// CalculateNarrationBudget computes available time for main narration.
func CalculateNarrationBudget(videoDuration int, disclaimerDuration float64) (float64, int) {
	musicTail := CalculateMusicTail(videoDuration)
	budgetSeconds := float64(videoDuration) - disclaimerDuration - musicTail

	// ~2.5 words per second
	budgetWords := int(budgetSeconds * 2.5)

	return budgetSeconds, budgetWords
}
