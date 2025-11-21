package prompts_test

import (
	"strings"
	"testing"

	"github.com/omnigen/backend/internal/prompts"
)

func TestPharmaceuticalAdGuidanceExists(t *testing.T) {
	if prompts.PharmaceuticalAdGuidance == "" {
		t.Error("PharmaceuticalAdGuidance should not be empty")
	}
}

func TestPharmaceuticalAdGuidanceContainsKeyElements(t *testing.T) {
	guidance := prompts.PharmaceuticalAdGuidance

	requiredElements := []string{
		"FDA",
		"benefits AND risks",
		"side effects",
		"disclaimer",
		"Consult your doctor",
		"adults 40-65",
		"responsible medication use",
	}

	for _, element := range requiredElements {
		if !strings.Contains(guidance, element) {
			t.Errorf("Pharmaceutical guidance should contain '%s'", element)
		}
	}
}

func TestPharmaceuticalAdGuidanceContainsSceneGuidance(t *testing.T) {
	guidance := prompts.PharmaceuticalAdGuidance

	// Ensure scene array planning guidance is included
	sceneElements := []string{
		"Scene Array Planning",
		"Early Scenes",
		"Middle Scenes",
		"Final Scenes",
		"generation_prompt",
	}

	for _, element := range sceneElements {
		if !strings.Contains(guidance, element) {
			t.Errorf("Pharmaceutical guidance should contain scene planning element '%s'", element)
		}
	}
}

func TestPharmaceuticalAdGuidanceContainsAntiRepetition(t *testing.T) {
	guidance := prompts.PharmaceuticalAdGuidance

	// Ensure anti-repetition guidance is included
	antiRepetitionElements := []string{
		"Scene Differentiation",
		"DIFFERENT location",
		"DIFFERENT action",
		"PROGRESSING emotional state",
	}

	for _, element := range antiRepetitionElements {
		if !strings.Contains(guidance, element) {
			t.Errorf("Pharmaceutical guidance should contain anti-repetition element '%s'", element)
		}
	}
}
