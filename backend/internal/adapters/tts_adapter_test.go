package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestOpenAITTSAdapter_VoiceMap(t *testing.T) {
	// Test that voice mapping is correct
	tests := []struct {
		inputVoice   string
		expectedVoice string
		shouldExist  bool
	}{
		{"male", "onyx", true},
		{"female", "nova", true},
		{"unknown", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.inputVoice, func(t *testing.T) {
			voice, ok := voiceMap[tt.inputVoice]
			if ok != tt.shouldExist {
				t.Errorf("voiceMap[%q] exists = %v, want %v", tt.inputVoice, ok, tt.shouldExist)
			}
			if ok && voice != tt.expectedVoice {
				t.Errorf("voiceMap[%q] = %v, want %v", tt.inputVoice, voice, tt.expectedVoice)
			}
		})
	}
}

func TestOpenAITTSAdapter_GenerateVoiceover_EmptyTextError(t *testing.T) {
	logger := zap.NewNop()
	adapter := NewOpenAITTSAdapter("test-api-key", logger)

	// GenerateVoiceover should handle empty text gracefully
	// The actual implementation may not check for empty, but GenerateVoiceoverWithDuration does
	_, _, err := adapter.GenerateVoiceoverWithDuration(context.Background(), "", "male", 1.0)
	if err == nil {
		t.Error("GenerateVoiceoverWithDuration should return error for empty text")
	}
	if !strings.Contains(err.Error(), "empty text") {
		t.Errorf("Error should mention 'empty text', got: %v", err)
	}
}

func TestOpenAITTSAdapter_GenerateVoiceover_InvalidVoice(t *testing.T) {
	logger := zap.NewNop()
	adapter := NewOpenAITTSAdapter("test-api-key", logger)

	// Test with invalid voice
	_, err := adapter.GenerateVoiceover(context.Background(), "test text", "invalid_voice")
	if err == nil {
		t.Error("GenerateVoiceover should return error for invalid voice")
	}
	if !strings.Contains(err.Error(), "invalid voice") {
		t.Errorf("Error should mention 'invalid voice', got: %v", err)
	}
}

func TestOpenAITTSAdapter_GenerateVoiceoverWithDuration_InvalidVoice(t *testing.T) {
	logger := zap.NewNop()
	adapter := NewOpenAITTSAdapter("test-api-key", logger)

	// Test with invalid voice
	_, _, err := adapter.GenerateVoiceoverWithDuration(context.Background(), "test text", "invalid_voice", 1.0)
	if err == nil {
		t.Error("GenerateVoiceoverWithDuration should return error for invalid voice")
	}
	if !strings.Contains(err.Error(), "invalid voice") {
		t.Errorf("Error should mention 'invalid voice', got: %v", err)
	}
}

func TestOpenAITTSAdapter_OpenAITTSRequest_Structure(t *testing.T) {
	// Test that the request struct has correct fields for the API
	req := openAITTSRequest{
		Model:          "tts-1",
		Input:          "Hello world",
		Voice:          "onyx",
		ResponseFormat: "mp3",
		Speed:          1.4,
	}

	if req.Model != "tts-1" {
		t.Errorf("Model = %v, want 'tts-1'", req.Model)
	}
	if req.Speed != 1.4 {
		t.Errorf("Speed = %v, want 1.4", req.Speed)
	}
	if req.ResponseFormat != "mp3" {
		t.Errorf("ResponseFormat = %v, want 'mp3'", req.ResponseFormat)
	}
}

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		status   int
		expected bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{429, false}, // Rate limit is not retryable in current implementation
		{500, true},
		{502, true},
		{503, true},
		{504, true},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			result := isRetryableStatus(tt.status)
			if result != tt.expected {
				t.Errorf("isRetryableStatus(%d) = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "retryable error",
			err:      &retryableError{err: context.DeadlineExceeded},
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRetryableError_Unwrap(t *testing.T) {
	originalErr := context.DeadlineExceeded
	wrappedErr := &retryableError{err: originalErr}

	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestRetryableError_Error(t *testing.T) {
	originalErr := context.DeadlineExceeded
	wrappedErr := &retryableError{err: originalErr}

	if wrappedErr.Error() != originalErr.Error() {
		t.Errorf("Error() = %v, want %v", wrappedErr.Error(), originalErr.Error())
	}
}

func TestNewOpenAITTSAdapter(t *testing.T) {
	logger := zap.NewNop()
	adapter := NewOpenAITTSAdapter("test-api-key", logger)

	if adapter == nil {
		t.Fatal("NewOpenAITTSAdapter should not return nil")
	}
	if adapter.apiKey != "test-api-key" {
		t.Errorf("apiKey = %v, want 'test-api-key'", adapter.apiKey)
	}
	if adapter.model != "tts-1" {
		t.Errorf("model = %v, want 'tts-1'", adapter.model)
	}
	if adapter.endpoint != "https://api.openai.com/v1/audio/speech" {
		t.Errorf("endpoint = %v, want 'https://api.openai.com/v1/audio/speech'", adapter.endpoint)
	}
	if adapter.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestOpenAITTSAdapter_CallOpenAITTS_NetworkError(t *testing.T) {
	logger := zap.NewNop()

	// Create a test server that immediately closes connections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate server error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	adapter := &OpenAITTSAdapter{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
		logger:     logger,
		model:      "tts-1",
		endpoint:   server.URL,
	}

	req := openAITTSRequest{
		Model:          "tts-1",
		Input:          "test text",
		Voice:          "onyx",
		ResponseFormat: "mp3",
		Speed:          1.0,
	}

	_, err := adapter.callOpenAITTS(context.Background(), req)
	if err == nil {
		t.Error("callOpenAITTS should return error for server error response")
	}

	// Should be wrapped as retryable error for 5xx
	if !isRetryableError(err) {
		t.Error("5xx error should be retryable")
	}
}

func TestOpenAITTSAdapter_CallOpenAITTS_ClientError(t *testing.T) {
	logger := zap.NewNop()

	// Create a test server that returns 4xx error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	adapter := &OpenAITTSAdapter{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
		logger:     logger,
		model:      "tts-1",
		endpoint:   server.URL,
	}

	req := openAITTSRequest{
		Model:          "tts-1",
		Input:          "test text",
		Voice:          "onyx",
		ResponseFormat: "mp3",
		Speed:          1.0,
	}

	_, err := adapter.callOpenAITTS(context.Background(), req)
	if err == nil {
		t.Error("callOpenAITTS should return error for client error response")
	}

	// Should NOT be retryable for 4xx
	if isRetryableError(err) {
		t.Error("4xx error should not be retryable")
	}
}

func TestOpenAITTSAdapter_GenerateVoiceover_SpeedParameter(t *testing.T) {
	// Verify that different speed values are accepted
	logger := zap.NewNop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return minimal valid audio data (just enough to not fail)
		w.WriteHeader(http.StatusOK)
		// Write some bytes that look like audio data header
		w.Write([]byte{0xFF, 0xFB, 0x90, 0x00})
	}))
	defer server.Close()

	adapter := &OpenAITTSAdapter{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
		logger:     logger,
		model:      "tts-1",
		endpoint:   server.URL,
	}

	req := openAITTSRequest{
		Model:          "tts-1",
		Input:          "test",
		Voice:          "onyx",
		ResponseFormat: "mp3",
		Speed:          1.4, // Disclaimer speed
	}

	data, err := adapter.callOpenAITTS(context.Background(), req)
	if err != nil {
		t.Errorf("callOpenAITTS with speed 1.4 should not fail: %v", err)
	}
	if len(data) == 0 {
		t.Error("callOpenAITTS should return data")
	}
}
