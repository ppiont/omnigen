package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// NonRetryableError wraps errors that should not be retried (e.g., 4xx client errors)
type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string { return e.Err.Error() }
func (e *NonRetryableError) Unwrap() error { return e.Err }

// NewNonRetryableError creates a new non-retryable error
func NewNonRetryableError(err error) *NonRetryableError {
	return &NonRetryableError{Err: err}
}

// IsNonRetryable checks if an error should not be retried
func IsNonRetryable(err error) bool {
	var nre *NonRetryableError
	return errors.As(err, &nre)
}

// Config holds retry configuration
type Config struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultConfig returns a sensible default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// APIConfig returns a retry configuration suitable for external API calls
// Uses longer delays to handle rate limiting and transient failures
func APIConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}
}

// Do executes the given function with exponential backoff retry logic
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry non-retryable errors (e.g., 4xx client errors)
		if IsNonRetryable(err) {
			return errors.Unwrap(err)
		}

		// Don't sleep after the last attempt
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt)))
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("max retries exceeded (%d attempts): %w", cfg.MaxAttempts, lastErr)
}
