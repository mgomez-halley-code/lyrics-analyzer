package client

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/service"
)

// RetryConfig holds configuration for the retry decorator
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
}

// RetryDecorator wraps a LyricsProvider with retry logic and exponential backoff
type RetryDecorator struct {
	client service.LyricsProvider
	config RetryConfig
}

// NewRetryDecorator creates a new retry decorator
func NewRetryDecorator(client service.LyricsProvider, config RetryConfig) *RetryDecorator {
	return &RetryDecorator{
		client: client,
		config: config,
	}
}

// Helper to centralize the retry loop logic
func (r *RetryDecorator) executeWithRetry(ctx context.Context, operation func() (any, error)) (any, error) {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		result, err := operation()
		if err == nil {
			return result, nil
		}

		if !r.shouldRetry(err) {
			return nil, err
		}

		lastErr = err
		if attempt < r.config.MaxRetries {
			backoff := r.calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				// Continue to next attempt
			}
		}
	}
	return nil, lastErr
}

// GetLyrics now uses the helper for a cleaner implementation
func (r *RetryDecorator) GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error) {
	res, err := r.executeWithRetry(ctx, func() (any, error) {
		return r.client.GetLyrics(ctx, track, artist)
	})
	if err != nil {
		return nil, err
	}
	return res.(*model.LyricsSourceData), nil
}

// RetryableError interface for errors that can indicate retryability
type RetryableError interface {
	error
	ShouldRetry() bool
}

// shouldRetry determines if an error is retryable
func (r *RetryDecorator) shouldRetry(err error) bool {
	// Check if error implements RetryableError interface
	var retryableErr RetryableError
	if errors.As(err, &retryableErr) {
		return retryableErr.ShouldRetry()
	}

	// Default: retry on network/unknown errors
	// Don't retry on wrapped context errors
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
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
