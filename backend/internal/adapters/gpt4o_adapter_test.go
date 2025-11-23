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
	// Note: In the two-pass system, narrator_script is NOT required in the first pass.
	// It will be generated separately with exact timing in the second pass.
	validPharmaScript := func() *domain.Script {
		script := validScript()
		// narrator_script can be empty - it's generated in the second pass
		script.AudioSpec.NarratorScript = ""
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
			name: "pharma without narrator_script passes (two-pass system generates it later)",
			script: func() *domain.Script {
				s := validPharmaScript()
				s.AudioSpec.NarratorScript = ""
				return s
			}(),
			requestedDur:     30,
			isPharmaceutical: true,
			wantErr:          false, // narrator_script no longer required in first pass
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

// --- Narration Generation Tests ---

func TestParseNarrationResponse(t *testing.T) {
	tests := []struct {
		name           string
		response       string
		expectedText   string
		expectedWords  int
		expectError    bool
	}{
		{
			name: "valid response with word count tag",
			response: `Living with chronic pain doesn't have to mean living a limited life.
With our new treatment, you can find relief and get back to doing what you love.
Ask your doctor if this treatment is right for you.

<word_count>38</word_count>`,
			expectedText:  "Living with chronic pain doesn't have to mean living a limited life.\nWith our new treatment, you can find relief and get back to doing what you love.\nAsk your doctor if this treatment is right for you.",
			expectedWords: 38,
			expectError:   false,
		},
		{
			name:           "response without word count tag calculates automatically",
			response:       "This is a simple test narration with exactly ten words here.",
			expectedText:   "This is a simple test narration with exactly ten words here.",
			expectedWords:  11, // Automatically counted: "This is a simple test narration with exactly ten words here." = 11 words
			expectError:    false,
		},
		{
			name:           "response with word count in middle of text",
			response:       "Some narration text. <word_count>5</word_count> More text after.",
			expectedText:   "Some narration text.  More text after.",
			expectedWords:  5, // Uses the tag value, not automatic count
			expectError:    false,
		},
		{
			name:           "empty response",
			response:       "",
			expectedText:   "",
			expectedWords:  0,
			expectError:    false,
		},
		{
			name:           "whitespace only response",
			response:       "   \n\t  ",
			expectedText:   "",
			expectedWords:  0,
			expectError:    false,
		},
		{
			name: "response with zero word count tag",
			response: `<word_count>0</word_count>`,
			expectedText:  "",
			expectedWords: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, wordCount, err := ParseNarrationResponse(tt.response)

			if (err != nil) != tt.expectError {
				t.Errorf("ParseNarrationResponse() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if text != tt.expectedText {
				t.Errorf("ParseNarrationResponse() text = %q, want %q", text, tt.expectedText)
			}

			if wordCount != tt.expectedWords {
				t.Errorf("ParseNarrationResponse() wordCount = %d, want %d", wordCount, tt.expectedWords)
			}
		})
	}
}

func TestParseNarrationResponse_WordCountAccuracy(t *testing.T) {
	// Test that automatic word counting is accurate
	tests := []struct {
		text          string
		expectedWords int
	}{
		{"one", 1},
		{"one two three", 3},
		{"one  two   three", 3}, // Multiple spaces
		{"one\ntwo\nthree", 3}, // Newlines
		{"one\ttwo\tthree", 3}, // Tabs
		{"", 0},                // Empty
		{"   ", 0},             // Whitespace only
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			_, wordCount, _ := ParseNarrationResponse(tt.text)
			if wordCount != tt.expectedWords {
				t.Errorf("Word count for %q = %d, want %d", tt.text, wordCount, tt.expectedWords)
			}
		})
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain JSON",
			input:    `{"title": "Test"}`,
			expected: `{"title": "Test"}`,
		},
		{
			name:     "JSON in markdown code block",
			input:    "```json\n{\"title\": \"Test\"}\n```",
			expected: `{"title": "Test"}`,
		},
		{
			name:     "JSON in plain code block",
			input:    "```\n{\"title\": \"Test\"}\n```",
			expected: `{"title": "Test"}`,
		},
		{
			name:     "no closing backticks",
			input:    "```json\n{\"title\": \"Test\"}",
			expected: `{"title": "Test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			// Trim whitespace for comparison since extraction may include newlines
			resultTrimmed := result
			expectedTrimmed := tt.expected
			for len(resultTrimmed) > 0 && (resultTrimmed[0] == '\n' || resultTrimmed[0] == ' ') {
				resultTrimmed = resultTrimmed[1:]
			}
			for len(resultTrimmed) > 0 && (resultTrimmed[len(resultTrimmed)-1] == '\n' || resultTrimmed[len(resultTrimmed)-1] == ' ') {
				resultTrimmed = resultTrimmed[:len(resultTrimmed)-1]
			}
			if resultTrimmed != expectedTrimmed {
				t.Errorf("extractJSON() = %q, want %q", resultTrimmed, expectedTrimmed)
			}
		})
	}
}

func TestBuildUserPrompt_PharmaceuticalAd(t *testing.T) {
	// Test that pharmaceutical ad prompts are built correctly
	req := &ScriptGenerationRequest{
		Prompt:      "Create a 30-second ad for Acme Pain Relief",
		Duration:    30,
		AspectRatio: "16:9",
		Voice:       "male",
		SideEffects: "May cause drowsiness and nausea. Do not take if allergic.",
	}

	prompt := buildUserPrompt(req)

	// Check that pharmaceutical-specific instructions are included
	if !containsSubstring(prompt, "PHARMACEUTICAL AD CONFIGURATION") {
		t.Error("Prompt should contain pharmaceutical ad configuration")
	}
	if !containsSubstring(prompt, "Do NOT generate narrator_script") {
		t.Error("Prompt should instruct not to generate narrator_script in first pass")
	}
	if !containsSubstring(prompt, "May cause drowsiness") {
		t.Error("Prompt should contain the side effects text")
	}
	if !containsSubstring(prompt, "STORE VERBATIM") {
		t.Error("Prompt should emphasize storing side effects verbatim")
	}
}

func TestBuildUserPrompt_NonPharmaceuticalAd(t *testing.T) {
	// Test that non-pharmaceutical ad prompts don't include pharma-specific instructions
	req := &ScriptGenerationRequest{
		Prompt:      "Create a 30-second ad for running shoes",
		Duration:    30,
		AspectRatio: "16:9",
		// No Voice or SideEffects
	}

	prompt := buildUserPrompt(req)

	// Check that pharmaceutical-specific instructions are NOT included
	if containsSubstring(prompt, "PHARMACEUTICAL AD CONFIGURATION") {
		t.Error("Non-pharma prompt should not contain pharmaceutical ad configuration")
	}
	if containsSubstring(prompt, "narrator_script") {
		t.Error("Non-pharma prompt should not mention narrator_script")
	}
}

func TestBuildUserPrompt_StartImage(t *testing.T) {
	req := &ScriptGenerationRequest{
		Prompt:      "Test ad",
		Duration:    15,
		AspectRatio: "9:16",
		StartImage:  "https://example.com/image.jpg",
	}

	prompt := buildUserPrompt(req)

	if !containsSubstring(prompt, "Starting Image") {
		t.Error("Prompt should mention starting image when provided")
	}
}

func TestMinMax(t *testing.T) {
	// Test the min/max helper functions
	tests := []struct {
		a, b int
		minExpected int
		maxExpected int
	}{
		{1, 2, 1, 2},
		{2, 1, 1, 2},
		{0, 0, 0, 0},
		{-1, 1, -1, 1},
		{100, 50, 50, 100},
	}

	for _, tt := range tests {
		if got := min(tt.a, tt.b); got != tt.minExpected {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.minExpected)
		}
		if got := max(tt.a, tt.b); got != tt.maxExpected {
			t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.maxExpected)
		}
	}
}
