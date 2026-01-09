package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockClient is a test mock that implements LyricsClient
type mockClient struct {
	callCount int
	responses []*LyricsData
	errors    []error
}

func (m *mockClient) GetLyrics(ctx context.Context, track, artist string) (*LyricsData, error) {
	if m.callCount >= len(m.responses) {
		return nil, errors.New("mock: no more responses configured")
	}

	response := m.responses[m.callCount]
	err := m.errors[m.callCount]
	m.callCount++

	return response, err
}

func TestRetryDecorator_SuccessFirstAttempt(t *testing.T) {
	t.Parallel()

	expectedLyrics := &LyricsData{TrackID: 123, TrackName: "Test Song"}
	mock := &mockClient{
		responses: []*LyricsData{expectedLyrics},
		errors:    []error{nil},
	}

	retryClient := NewRetryDecorator(mock, DefaultRetryConfig())
	lyrics, err := retryClient.GetLyrics(context.Background(), "Test Song", "Test Artist")

	assert.NoError(t, err)
	assert.Equal(t, expectedLyrics, lyrics)
	assert.Equal(t, 1, mock.callCount)
}

func TestRetryDecorator_ServerErrorThenSuccess(t *testing.T) {
	t.Parallel()

	expectedLyrics := &LyricsData{TrackID: 123}
	mock := &mockClient{
		responses: []*LyricsData{nil, nil, expectedLyrics},
		errors: []error{
			&APIError{StatusCode: 500},
			&APIError{StatusCode: 503},
			nil,
		},
	}

	config := RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
	}

	retryClient := NewRetryDecorator(mock, config)
	lyrics, err := retryClient.GetLyrics(context.Background(), "Test Song", "Test Artist")

	assert.NoError(t, err)
	assert.NotNil(t, lyrics)
	assert.Equal(t, 3, mock.callCount)
}

func TestRetryDecorator_NoRetryOnNotFound(t *testing.T) {
	t.Parallel()

	mock := &mockClient{
		responses: []*LyricsData{nil, nil},
		errors:    []error{ErrLyricsNotFound, nil},
	}

	retryClient := NewRetryDecorator(mock, DefaultRetryConfig())
	_, err := retryClient.GetLyrics(context.Background(), "Test Song", "Test Artist")

	assert.ErrorIs(t, err, ErrLyricsNotFound)
	assert.Equal(t, 1, mock.callCount, "Should not retry on 404/Sentinel error")
}

func TestRetryDecorator_ExhaustsRetries(t *testing.T) {
	t.Parallel()

	serverError := &APIError{StatusCode: 500}
	mock := &mockClient{
		responses: []*LyricsData{nil, nil, nil, nil},
		errors:    []error{serverError, serverError, serverError, serverError},
	}

	config := RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
	}

	retryClient := NewRetryDecorator(mock, config)
	_, err := retryClient.GetLyrics(context.Background(), "Test Song", "Test Artist")

	assert.Error(t, err)
	assert.Equal(t, 4, mock.callCount, "Initial call + 3 retries")
}

func TestRetryDecorator_ContextCancellation(t *testing.T) {
	t.Parallel()

	mock := &mockClient{
		responses: []*LyricsData{nil, nil, nil},
		errors:    []error{&APIError{StatusCode: 500}, &APIError{StatusCode: 500}, &APIError{StatusCode: 500}},
	}

	config := RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     2 * time.Second,
		Multiplier:     2.0,
	}

	retryClient := NewRetryDecorator(mock, config)

	// Set a very short timeout. The first call happens immediately,
	// then it hits the backoff and should be cancelled there.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := retryClient.GetLyrics(ctx, "Test", "Artist")

	// We check for DeadlineExceeded because the select block in your
	// decorator returns ctx.Err()
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// It should have only called the mock ONCE.
	assert.Equal(t, 1, mock.callCount, "Should stop during first backoff sleep")
}
