package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TTSAdapter defines the interface for text-to-speech generation.
type TTSAdapter interface {
	// GenerateVoiceover generates speech audio at 1.0x speed.
	// Variable speed (e.g. 1.4x for side effects) is applied post-generation using ffmpeg.
	GenerateVoiceover(ctx context.Context, text string, voice string) ([]byte, error)

	// GenerateVoiceoverWithDuration generates speech audio and returns duration.
	// Speed parameter allows 1.4x for side effects disclaimers.
	GenerateVoiceoverWithDuration(ctx context.Context, text string, voice string, speed float64) ([]byte, float64, error)
}

// OpenAITTSAdapter implements text-to-speech using the OpenAI TTS API.
type OpenAITTSAdapter struct {
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
	model      string
	endpoint   string
}

// NewOpenAITTSAdapter creates a new OpenAI TTS adapter.
func NewOpenAITTSAdapter(apiKey string, logger *zap.Logger) *OpenAITTSAdapter {
	return &OpenAITTSAdapter{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger:   logger,
		model:    "tts-1",
		endpoint: "https://api.openai.com/v1/audio/speech",
	}
}

// voiceMap provides the OpenAI voice IDs for supported narrator voices.
var voiceMap = map[string]string{
	"male":   "onyx",
	"female": "nova",
}

// openAITTSRequest matches the OpenAI TTS API schema.
type openAITTSRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format"`
	Speed          float64 `json:"speed"`
}

// retryableError wraps an error that should trigger a retry.
type retryableError struct {
	err error
}

func (e *retryableError) Error() string {
	return e.err.Error()
}

func (e *retryableError) Unwrap() error {
	return e.err
}

// GenerateVoiceover generates speech audio from the provided text using the configured voice.
func (t *OpenAITTSAdapter) GenerateVoiceover(ctx context.Context, text string, voice string) ([]byte, error) {
	startTime := time.Now()

	t.logger.Info("Generating voiceover with OpenAI TTS",
		zap.String("voice", voice),
		zap.Int("text_length", len(text)),
		zap.String("model", t.model),
	)

	openAIVoice, ok := voiceMap[voice]
	if !ok {
		return nil, fmt.Errorf("invalid voice selection: %s (expected 'male' or 'female')", voice)
	}

	reqPayload := openAITTSRequest{
		Model:          t.model,
		Input:          text,
		Voice:          openAIVoice,
		ResponseFormat: "mp3",
		Speed:          1.0,
	}

	attempts := 3
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	var lastErr error

	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			t.logger.Warn("Retrying OpenAI TTS request",
				zap.Int("attempt", attempt+1),
				zap.Int("max_attempts", attempts),
				zap.Error(lastErr),
			)

			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("tts retry cancelled: %w", ctx.Err())
			case <-time.After(delays[attempt-1]):
				// continue
			}
		}

		audioData, err := t.callOpenAITTS(ctx, reqPayload)
		if err == nil {
			duration := time.Since(startTime)
			t.logger.Info("Generated voiceover with OpenAI TTS",
				zap.String("voice", voice),
				zap.Int("audio_size_bytes", len(audioData)),
				zap.Duration("duration", duration),
				zap.Int("attempt", attempt+1),
			)
			return audioData, nil
		}

		lastErr = err

		if !isRetryableError(err) {
			t.logger.Error("OpenAI TTS request failed with non-retryable error",
				zap.Error(err),
				zap.Int("attempt", attempt+1),
			)
			return nil, fmt.Errorf("tts generation failed: %w", err)
		}
	}

	t.logger.Error("OpenAI TTS generation failed after maximum retries",
		zap.Error(lastErr),
		zap.Int("attempts", attempts),
	)
	return nil, fmt.Errorf("tts generation failed after %d attempts: %w", attempts, lastErr)
}

func (t *OpenAITTSAdapter) callOpenAITTS(ctx context.Context, reqPayload openAITTSRequest) ([]byte, error) {
	payload, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TTS request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+t.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	t.logger.Debug("Calling OpenAI TTS API",
		zap.String("endpoint", t.endpoint),
		zap.String("voice", reqPayload.Voice),
		zap.String("model", reqPayload.Model),
	)

	resp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return nil, &retryableError{err: fmt.Errorf("network error calling OpenAI TTS: %w", err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &retryableError{err: fmt.Errorf("failed to read TTS response: %w", err)}
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := fmt.Errorf("openai tts error (status %d): %s", resp.StatusCode, string(body))

		if isRetryableStatus(resp.StatusCode) {
			return nil, &retryableError{err: apiErr}
		}

		return nil, apiErr
	}

	return body, nil
}

func isRetryableError(err error) bool {
	var rerr *retryableError
	return errors.As(err, &rerr)
}

func isRetryableStatus(status int) bool {
	// Retry on server errors
	if status >= 500 && status < 600 {
		return true
	}
	// Do not retry on client errors (4xx), including 429 rate limit.
	return false
}

// GenerateVoiceoverWithDuration generates TTS audio at specified speed and returns duration.
func (t *OpenAITTSAdapter) GenerateVoiceoverWithDuration(ctx context.Context, text string, voice string, speed float64) ([]byte, float64, error) {
	if text == "" {
		return nil, 0, fmt.Errorf("empty text for TTS")
	}

	openAIVoice, ok := voiceMap[voice]
	if !ok {
		return nil, 0, fmt.Errorf("invalid voice selection: %s (expected 'male' or 'female')", voice)
	}

	reqPayload := openAITTSRequest{
		Model:          t.model,
		Input:          text,
		Voice:          openAIVoice,
		ResponseFormat: "mp3",
		Speed:          speed,
	}

	t.logger.Info("Generating voiceover with duration",
		zap.String("voice", voice),
		zap.Int("text_length", len(text)),
		zap.Float64("speed", speed),
	)

	audioData, err := t.callOpenAITTS(ctx, reqPayload)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to generate TTS: %w", err)
	}

	// Write to temp file to get duration via ffprobe
	tmpFile, err := os.CreateTemp("", "tts-*.mp3")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	defer func() { _ = tmpFile.Close() }()

	if _, err := tmpFile.Write(audioData); err != nil {
		return nil, 0, fmt.Errorf("failed to write temp file: %w", err)
	}

	duration, err := getAudioDurationFromFile(tmpFile.Name())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get audio duration: %w", err)
	}

	t.logger.Info("Generated voiceover with duration",
		zap.Int("audio_size_bytes", len(audioData)),
		zap.Float64("duration_seconds", duration),
	)

	return audioData, duration, nil
}

// getAudioDurationFromFile uses ffprobe to get audio duration in seconds.
func getAudioDurationFromFile(path string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}
	return strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
}
