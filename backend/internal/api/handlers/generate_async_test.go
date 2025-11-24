package handlers

import (
	"testing"
)

func TestBuildFadeFilters(t *testing.T) {
	tests := []struct {
		name           string
		fadeIn         float64
		fadeOut        float64
		totalDuration  float64
		expectedFilter string
	}{
		{
			name:           "standard 30s video",
			fadeIn:         1.5,
			fadeOut:        2.0,
			totalDuration:  30.0,
			expectedFilter: "fade=t=in:st=0:d=1.50,fade=t=out:st=28.00:d=2.00",
		},
		{
			name:           "short 5s video - both fades fit",
			fadeIn:         1.5,
			fadeOut:        2.0,
			totalDuration:  5.0,
			expectedFilter: "fade=t=in:st=0:d=1.50,fade=t=out:st=3.00:d=2.00",
		},
		{
			name:           "very short video - fade out only",
			fadeIn:         1.5,
			fadeOut:        2.0,
			totalDuration:  3.0,
			expectedFilter: "fade=t=out:st=0.00:d=2.00",
		},
		{
			name:           "15s video",
			fadeIn:         1.5,
			fadeOut:        2.0,
			totalDuration:  15.0,
			expectedFilter: "fade=t=in:st=0:d=1.50,fade=t=out:st=13.00:d=2.00",
		},
		{
			name:           "60s video",
			fadeIn:         1.5,
			fadeOut:        2.0,
			totalDuration:  60.0,
			expectedFilter: "fade=t=in:st=0:d=1.50,fade=t=out:st=58.00:d=2.00",
		},
		{
			name:           "exact boundary - fades just fit",
			fadeIn:         1.5,
			fadeOut:        2.0,
			totalDuration:  3.5,
			expectedFilter: "fade=t=in:st=0:d=1.50,fade=t=out:st=1.50:d=2.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFadeFilters(tt.fadeIn, tt.fadeOut, tt.totalDuration)
			if result != tt.expectedFilter {
				t.Errorf("buildFadeFilters(%.1f, %.1f, %.1f) = %s, expected %s",
					tt.fadeIn, tt.fadeOut, tt.totalDuration, result, tt.expectedFilter)
			}
		})
	}
}

func TestGetVideoDuration(t *testing.T) {
	// Test with a non-existent file - should return 0 gracefully
	t.Run("non-existent file returns 0", func(t *testing.T) {
		result := getVideoDuration("/non/existent/file.mp4")
		if result != 0 {
			t.Errorf("getVideoDuration for non-existent file = %f, expected 0", result)
		}
	})

	// Test with an invalid path - should return 0 gracefully
	t.Run("invalid path returns 0", func(t *testing.T) {
		result := getVideoDuration("")
		if result != 0 {
			t.Errorf("getVideoDuration for empty path = %f, expected 0", result)
		}
	})
}

func TestS3KeyGenerationHelpers(t *testing.T) {
	userID := "user123"
	jobID := "job456"

	testCases := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "scene clip key",
			got:  buildSceneClipKey(userID, jobID, 3),
			want: "users/user123/jobs/job456/clips/scene-003.mp4",
		},
		{
			name: "scene thumbnail key",
			got:  buildSceneThumbnailKey(userID, jobID, 5),
			want: "users/user123/jobs/job456/thumbnails/scene-005.jpg",
		},
		{
			name: "job thumbnail key",
			got:  buildJobThumbnailKey(userID, jobID),
			want: "users/user123/jobs/job456/thumbnails/job-thumbnail.jpg",
		},
		{
			name: "background music key",
			got:  buildAudioKey(userID, jobID),
			want: "users/user123/jobs/job456/audio/background-music.mp3",
		},
		{
			name: "narrator audio key",
			got:  buildNarratorAudioKey(userID, jobID),
			want: "users/user123/jobs/job456/audio/narrator-voiceover.mp3",
		},
		{
			name: "final video key",
			got:  buildFinalVideoKey(userID, jobID),
			want: "users/user123/jobs/job456/final/video.mp4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, tc.got)
			}
		})
	}
}

func TestEscapeFFmpegText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text - no escaping needed",
			input:    "Simple text",
			expected: "Simple text",
		},
		{
			name:     "text with colon",
			input:    "Text with: colon",
			expected: "Text with\\: colon",
		},
		{
			name:     "text with single quote",
			input:    "It's a test",
			expected: "It'\\''s a test",
		},
		{
			name:     "text with backslash",
			input:    "Back\\slash",
			expected: "Back\\\\slash",
		},
		{
			name:     "text with multiple special chars",
			input:    "It's got: everything\\here",
			expected: "It'\\''s got\\: everything\\\\here",
		},
		{
			name:     "CTA text example",
			input:    "Ask your doctor if ProductX is right for you",
			expected: "Ask your doctor if ProductX is right for you",
		},
		{
			name:     "side effects with colon",
			input:    "Side effects may include: headache, nausea",
			expected: "Side effects may include\\: headache, nausea",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeFFmpegText(tt.input)
			if result != tt.expected {
				t.Errorf("escapeFFmpegText(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
