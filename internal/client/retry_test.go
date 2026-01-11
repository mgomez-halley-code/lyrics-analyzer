package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
	"github.com/stretchr/testify/assert"
)

// mockAPIError is a test error that implements RetryableError
type mockAPIError struct {
	statusCode int
	message    string
}

func (e *mockAPIError) Error() string {
	return e.message
}

func (e *mockAPIError) ShouldRetry() bool {
	return e.statusCode >= 500
}

// mockNotFoundError simulates a non-retryable error that implements RetryableError
type mockNotFoundError struct{}

func (e *mockNotFoundError) Error() string {
	return "mock: lyrics not found"
}

func (e *mockNotFoundError) ShouldRetry() bool {
	return false
}

// mockClient is a test mock that implements LyricsClient
type mockClient struct {
	callCount int
	responses []*model.LyricsSourceData
	errors    []error
}

func (m *mockClient) GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error) {
	if m.callCount >= len(m.responses) {
		return nil, errors.New("mock: no more responses configured")
	}

	response := m.responses[m.callCount]
	err := m.errors[m.callCount]
	m.callCount++

	return response, err
}

func TestRetryDecorator_GetLyrics(t *testing.T) {
	tests := []struct {
		name             string
		responses        []*model.LyricsSourceData
		errors           []error
		maxRetries       int
		initialBackoff   time.Duration
		expectedCalls    int
		shouldError      bool
		expectedResponse *model.LyricsSourceData
	}{
		{
			name:             "success first attempt",
			responses:        []*model.LyricsSourceData{{TrackID: 123, TrackName: "Test Song"}},
			errors:           []error{nil},
			maxRetries:       3,
			initialBackoff:   10 * time.Millisecond,
			expectedCalls:    1,
			shouldError:      false,
			expectedResponse: &model.LyricsSourceData{TrackID: 123, TrackName: "Test Song"},
		},
		{
			name: "retry 2 times then success",
			responses: []*model.LyricsSourceData{
				nil,
				nil,
				{TrackID: 123},
			},
			errors: []error{
				&mockAPIError{statusCode: 500, message: "internal server error"},
				&mockAPIError{statusCode: 503, message: "service unavailable"},
				nil,
			},
			maxRetries:       3,
			initialBackoff:   10 * time.Millisecond,
			expectedCalls:    3,
			shouldError:      false,
			expectedResponse: &model.LyricsSourceData{TrackID: 123},
		},
		{
			name:           "no retry on 404 not found",
			responses:      []*model.LyricsSourceData{nil, nil},
			errors:         []error{&mockNotFoundError{}, nil},
			maxRetries:     3,
			initialBackoff: 10 * time.Millisecond,
			expectedCalls:  1,
			shouldError:    true,
		},
		{
			name: "exhaust all retries",
			responses: []*model.LyricsSourceData{
				nil,
				nil,
				nil,
				nil,
			},
			errors: []error{
				&mockAPIError{statusCode: 500, message: "internal server error"},
				&mockAPIError{statusCode: 500, message: "internal server error"},
				&mockAPIError{statusCode: 500, message: "internal server error"},
				&mockAPIError{statusCode: 500, message: "internal server error"},
			},
			maxRetries:     3,
			initialBackoff: 1 * time.Millisecond,
			expectedCalls:  4, // Initial call + 3 retries
			shouldError:    true,
		},
		{
			name: "different 5xx errors all retry",
			responses: []*model.LyricsSourceData{
				nil,
				nil,
				nil,
				{TrackID: 456},
			},
			errors: []error{
				&mockAPIError{statusCode: 500, message: "internal server error"},
				&mockAPIError{statusCode: 502, message: "bad gateway"},
				&mockAPIError{statusCode: 503, message: "service unavailable"},
				nil,
			},
			maxRetries:       3,
			initialBackoff:   1 * time.Millisecond,
			expectedCalls:    4,
			shouldError:      false,
			expectedResponse: &model.LyricsSourceData{TrackID: 456},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockClient{
				responses: tt.responses,
				errors:    tt.errors,
			}

			config := RetryConfig{
				MaxRetries:     tt.maxRetries,
				InitialBackoff: tt.initialBackoff,
				MaxBackoff:     100 * time.Millisecond,
				Multiplier:     2.0,
			}

			retryClient := NewRetryDecorator(mock, config)
			lyrics, err := retryClient.GetLyrics(context.Background(), "Test Song", "Test Artist")

			assert.Equal(t, tt.expectedCalls, mock.callCount)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedResponse != nil {
					assert.Equal(t, tt.expectedResponse, lyrics)
				} else {
					assert.NotNil(t, lyrics)
				}
			}
		})
	}
}

func TestRetryDecorator_ContextCancellation(t *testing.T) {
	t.Parallel()

	mock := &mockClient{
		responses: []*model.LyricsSourceData{nil, nil, nil},
		errors: []error{
			&mockAPIError{statusCode: 500, message: "internal server error"},
			&mockAPIError{statusCode: 500, message: "internal server error"},
			&mockAPIError{statusCode: 500, message: "internal server error"},
		},
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
