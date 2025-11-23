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
		"Consult your doctor",
		"Responsible medication use", // Updated casing to match new guidance
	}

	for _, element := range requiredElements {
		if !strings.Contains(guidance, element) {
			t.Errorf("Pharmaceutical guidance should contain '%s'", element)
		}
	}
}

func TestPharmaceuticalAdGuidanceContainsSceneGuidance(t *testing.T) {
	guidance := prompts.PharmaceuticalAdGuidance

	// Ensure patient journey arc and scene planning guidance is included
	// Updated to match new enhanced guidance structure
	sceneElements := []string{
		"PATIENT JOURNEY ARC",
		"Early scenes",
		"Middle scenes",
		"Later scenes",
		"Final scenes",
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
	// Updated to match new 5-dimension progression system
	antiRepetitionElements := []string{
		"5-DIMENSION PROGRESSION",
		"ANTI-REPETITION",
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

func TestPharmaceuticalAdGuidanceContainsVisualConstants(t *testing.T) {
	guidance := prompts.PharmaceuticalAdGuidance

	// Ensure visual constants extraction guidance is included (new feature)
	visualConstantElements := []string{
		"VISUAL CONSTANTS EXTRACTION",
		"PATIENT ARCHETYPE",
		"CONDITION VISUALIZATION",
		"BRAND COLOR PALETTE",
		"MEDICATION VISUAL TREATMENT",
		"LIGHTING/MOOD PROGRESSION",
	}

	for _, element := range visualConstantElements {
		if !strings.Contains(guidance, element) {
			t.Errorf("Pharmaceutical guidance should contain visual constants element '%s'", element)
		}
	}
}

func TestPharmaceuticalAdGuidanceContainsSceneSpecificity(t *testing.T) {
	guidance := prompts.PharmaceuticalAdGuidance

	// Ensure scene specificity requirements are included (new feature)
	specificityElements := []string{
		"SCENE SPECIFICITY REQUIREMENTS",
		"NAMED PATIENT",
		"SPECIFIC LOCATIONS",
		"UNIQUE ACTION VERBS",
		"CONCRETE COLORS/LIGHTING",
		"MEASURABLE CAMERA",
	}

	for _, element := range specificityElements {
		if !strings.Contains(guidance, element) {
			t.Errorf("Pharmaceutical guidance should contain scene specificity element '%s'", element)
		}
	}
}
