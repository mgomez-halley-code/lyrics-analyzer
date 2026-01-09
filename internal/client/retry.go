package client

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds configuration for the retry decorator
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
}

// DefaultRetryConfig returns sensible defaults for retry behavior
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2.0,
	}
}

// RetryDecorator wraps a LyricsClient with retry logic and exponential backoff
type RetryDecorator struct {
	client LyricsClient
	config RetryConfig
}

// NewRetryDecorator creates a new retry decorator
func NewRetryDecorator(client LyricsClient, config RetryConfig) *RetryDecorator {
	return &RetryDecorator{
		client: client,
		config: config,
	}
}

// GetLyrics implements LyricsClient with retry logic
func (r *RetryDecorator) GetLyrics(ctx context.Context, track, artist string) (*LyricsData, error) {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		lyrics, err := r.client.GetLyrics(ctx, track, artist)

		if err == nil {
			return lyrics, nil
		}

		// Don't retry on certain errors
		if !r.shouldRetry(err) {
			return nil, err
		}

		lastErr = err

		if attempt < r.config.MaxRetries {
			backoff := r.calculateBackoff(attempt)

			// Check if context is already cancelled
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				// Continue to next attempt
			}
		}
	}

	// All retries exhausted
	return nil, lastErr
}

// shouldRetry determines if an error is retryable
func (r *RetryDecorator) shouldRetry(err error) bool {
	// Don't retry client-side errors
	if errors.Is(err, ErrLyricsNotFound) {
		return false
	}

	// Retry on server errors (500, 503) and network errors
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500
	}

	return true
}

// calculateBackoff calculates the backoff duration for a given attempt
func (r *RetryDecorator) calculateBackoff(attempt int) time.Duration {
	// 1. Calculate the base exponential backoff: initialBackoff * (multiplier ^ attempt)
	backoff := float64(r.config.InitialBackoff) * math.Pow(r.config.Multiplier, float64(attempt))

	// 2. Cap at max backoff
	if backoff > float64(r.config.MaxBackoff) {
		backoff = float64(r.config.MaxBackoff)
	}

	// 3. Apply Jitter
	// We use the calculated backoff as the maximum range for a random duration.
	// This is the "Full Jitter" strategy recommended by AWS for high-scale systems.
	nanos := int64(backoff)
	if nanos <= 0 {
		return r.config.InitialBackoff
	}

	jitteredNanos := rand.Int63n(nanos)
	return time.Duration(jitteredNanos)
}
