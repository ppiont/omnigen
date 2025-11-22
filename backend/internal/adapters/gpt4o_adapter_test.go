package adapters

import (
	"testing"

	"github.com/omnigen/backend/internal/domain"
)

func TestValidateScript(t *testing.T) {
	// Helper to create a valid base script
	validScript := func() *domain.Script {
		return &domain.Script{
			Title:         "Test Ad",
			TotalDuration: 30,
			Scenes: []domain.Scene{
				{
					SceneNumber:      1,
					GenerationPrompt: "A detailed scene description with plenty of visual information for generation",
					Duration:         10,
				},
				{
					SceneNumber:      2,
					GenerationPrompt: "Another detailed scene description with enough characters for validation",
					Duration:         10,
				},
				{
					SceneNumber:      3,
					GenerationPrompt: "Final scene with comprehensive visual details and camera directions included",
					Duration:         10,
				},
			},
			AudioSpec: domain.AudioSpec{
				MusicMood:  "upbeat",
				MusicStyle: "electronic",
			},
		}
	}

	// Helper to create a valid pharma script
	validPharmaScript := func() *domain.Script {
		script := validScript()
		script.AudioSpec.NarratorScript = "A comprehensive narrator script explaining the medication benefits and side effects"
		script.AudioSpec.SideEffectsText = "May cause drowsiness. Consult your doctor before use."
		return script
	}

	tests := []struct {
		name             string
		script           *domain.Script
		requestedDur     int
		isPharmaceutical bool
		wantErr          bool
		errContains      string
	}{
		{
			name:             "valid script passes",
			script:           validScript(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          false,
		},
		{
			name: "empty title fails",
			script: func() *domain.Script {
				s := validScript()
				s.Title = ""
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "title is empty",
		},
		{
			name: "no scenes fails",
			script: func() *domain.Script {
				s := validScript()
				s.Scenes = []domain.Scene{}
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "no scenes",
		},
		{
			name:             "duration mismatch fails",
			script:           validScript(),
			requestedDur:     60, // Script has 30, requested 60
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "doesn't match requested",
		},
		{
			name: "incorrect scene number fails",
			script: func() *domain.Script {
				s := validScript()
				s.Scenes[1].SceneNumber = 5 // Should be 2
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "incorrect scene_number",
		},
		{
			name: "empty generation_prompt fails",
			script: func() *domain.Script {
				s := validScript()
				s.Scenes[0].GenerationPrompt = ""
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "empty generation_prompt",
		},
		{
			name: "short generation_prompt fails",
			script: func() *domain.Script {
				s := validScript()
				s.Scenes[0].GenerationPrompt = "Too short"
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "suspiciously short generation_prompt",
		},
		{
			name: "placeholder [insert text fails",
			script: func() *domain.Script {
				s := validScript()
				s.Scenes[0].GenerationPrompt = "A scene with [insert description here] and more visual details for the prompt"
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "placeholder text",
		},
		{
			name: "placeholder [placeholder text fails",
			script: func() *domain.Script {
				s := validScript()
				s.Scenes[1].GenerationPrompt = "A scene with [placeholder for visual description] and additional content"
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "placeholder text",
		},
		{
			name: "placeholder [TBD text fails",
			script: func() *domain.Script {
				s := validScript()
				s.Scenes[2].GenerationPrompt = "A scene with [TBD - add details later] and some additional content here"
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "placeholder text",
		},
		{
			name: "missing music_mood fails",
			script: func() *domain.Script {
				s := validScript()
				s.AudioSpec.MusicMood = ""
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "missing music_mood",
		},
		{
			name: "missing music_style fails",
			script: func() *domain.Script {
				s := validScript()
				s.AudioSpec.MusicStyle = ""
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          true,
			errContains:      "missing music_style",
		},
		// Pharmaceutical-specific tests
		{
			name:             "valid pharma script passes",
			script:           validPharmaScript(),
			requestedDur:     30,
			isPharmaceutical: true,
			wantErr:          false,
		},
		{
			name: "pharma without narrator_script fails",
			script: func() *domain.Script {
				s := validPharmaScript()
				s.AudioSpec.NarratorScript = ""
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: true,
			wantErr:          true,
			errContains:      "missing narrator_script",
		},
		{
			name: "pharma without side_effects_text fails",
			script: func() *domain.Script {
				s := validPharmaScript()
				s.AudioSpec.SideEffectsText = ""
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: true,
			wantErr:          true,
			errContains:      "missing side_effects_text",
		},
		{
			name: "non-pharma script without narrator_script passes",
			script: func() *domain.Script {
				s := validScript()
				// No narrator script, but not pharmaceutical
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: false,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScript(tt.script, tt.requestedDur, tt.isPharmaceutical)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("validateScript() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

// contains checks if s contains substr (case-insensitive for flexibility)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
