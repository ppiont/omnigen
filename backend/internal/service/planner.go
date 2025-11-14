package service

import (
	"fmt"
	"strings"

	"github.com/omnigen/backend/internal/domain"
	"go.uber.org/zap"
)

// ScenePlanner plans video scenes based on parsed prompts
type ScenePlanner struct {
	logger *zap.Logger
}

// NewScenePlanner creates a new scene planner
func NewScenePlanner(logger *zap.Logger) *ScenePlanner {
	return &ScenePlanner{
		logger: logger,
	}
}

// PlanScenes creates a scene plan for ad creative
func (p *ScenePlanner) PlanScenes(parsed *domain.ParsedPrompt) ([]domain.Scene, error) {
	duration := float64(parsed.Duration)

	// For ad creative, use 3-scene structure
	var scenes []domain.Scene

	if duration <= 30 {
		// Short ad (15-30s): 3 scenes
		scenes = p.planShortAd(parsed, duration)
	} else {
		// Longer ad (30-60s+): 4 scenes
		scenes = p.planLongAd(parsed, duration)
	}

	p.logger.Info("Scenes planned successfully",
		zap.Int("scene_count", len(scenes)),
		zap.Float64("total_duration", duration),
	)

	return scenes, nil
}

// planShortAd creates a 3-scene plan for short ads (15-30s)
func (p *ScenePlanner) planShortAd(parsed *domain.ParsedPrompt, totalDuration float64) []domain.Scene {
	styleString := strings.Join(parsed.VisualStyle, ", ")
	colorString := strings.Join(parsed.ColorPalette, ", ")

	// Scene 1: Product intro (30% of duration)
	scene1Duration := totalDuration * 0.30
	scene1 := domain.Scene{
		Number:   1,
		Duration: scene1Duration,
		Prompt: fmt.Sprintf("%s emerging from elegant background, %s style, %s colors, dramatic reveal",
			parsed.ProductType, styleString, colorString),
		Style:      styleString,
		Transition: "fade",
	}

	// Scene 2: Product showcase (40% of duration)
	scene2Duration := totalDuration * 0.40
	scene2 := domain.Scene{
		Number:   2,
		Duration: scene2Duration,
		Prompt: fmt.Sprintf("%s in detail, highlighting features, %s aesthetic, %s tones, professional lighting",
			parsed.ProductType, styleString, colorString),
		Style:      styleString,
		Transition: "dissolve",
	}

	// Scene 3: Brand/CTA (30% of duration)
	scene3Duration := totalDuration * 0.30
	cta := "shop now"
	if len(parsed.TextOverlays) > 0 {
		cta = parsed.TextOverlays[0]
	}

	scene3 := domain.Scene{
		Number:   3,
		Duration: scene3Duration,
		Prompt: fmt.Sprintf("%s with brand logo, \"%s\" text overlay, %s style, %s background, call to action",
			parsed.ProductType, cta, styleString, colorString),
		Style:      styleString,
		Transition: "fade",
	}

	return []domain.Scene{scene1, scene2, scene3}
}

// planLongAd creates a 4-scene plan for longer ads (30-60s+)
func (p *ScenePlanner) planLongAd(parsed *domain.ParsedPrompt, totalDuration float64) []domain.Scene {
	styleString := strings.Join(parsed.VisualStyle, ", ")
	colorString := strings.Join(parsed.ColorPalette, ", ")

	// Scene 1: Hook (20% of duration)
	scene1Duration := totalDuration * 0.20
	scene1 := domain.Scene{
		Number:   1,
		Duration: scene1Duration,
		Prompt: fmt.Sprintf("cinematic opening shot, %s atmosphere, %s colors, attention-grabbing",
			styleString, colorString),
		Style:      styleString,
		Transition: "fade",
	}

	// Scene 2: Product reveal (25% of duration)
	scene2Duration := totalDuration * 0.25
	scene2 := domain.Scene{
		Number:   2,
		Duration: scene2Duration,
		Prompt: fmt.Sprintf("%s reveal, %s style, %s tones, dramatic presentation",
			parsed.ProductType, styleString, colorString),
		Style:      styleString,
		Transition: "cut",
	}

	// Scene 3: Product showcase (30% of duration)
	scene3Duration := totalDuration * 0.30
	scene3 := domain.Scene{
		Number:   3,
		Duration: scene3Duration,
		Prompt: fmt.Sprintf("%s in use, detailed features, %s aesthetic, %s colors, lifestyle context",
			parsed.ProductType, styleString, colorString),
		Style:      styleString,
		Transition: "dissolve",
	}

	// Scene 4: CTA/Brand (25% of duration)
	scene4Duration := totalDuration * 0.25
	cta := "discover more"
	if len(parsed.TextOverlays) > 0 {
		cta = parsed.TextOverlays[0]
	}

	scene4 := domain.Scene{
		Number:   4,
		Duration: scene4Duration,
		Prompt: fmt.Sprintf("%s with brand identity, \"%s\" message, %s style, powerful closing",
			parsed.ProductType, cta, styleString),
		Style:      styleString,
		Transition: "fade",
	}

	return []domain.Scene{scene1, scene2, scene3, scene4}
}
