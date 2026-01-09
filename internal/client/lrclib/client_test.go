package lrclib

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetLyrics_Logic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectError    bool
		expectSentinel error
		errorContains  string
	}{
		{
			name:         "Success - 200 OK with results",
			statusCode:   http.StatusOK,
			responseBody: `[{"id":123,"trackName":"Song","artistName":"Artist","duration":180.0}]`,
			expectError:  false,
		},
		{
			name:           "Not Found - Empty search results",
			statusCode:     http.StatusOK,
			responseBody:   `[]`,
			expectError:    true,
			expectSentinel: ErrLyricsNotFound,
		},
		{
			name:           "Not Found - 404 Status",
			statusCode:     http.StatusNotFound,
			responseBody:   `Not Found`,
			expectError:    true,
			expectSentinel: ErrLyricsNotFound,
		},
		{
			name:          "Parsing Error - Malformed JSON",
			statusCode:    http.StatusOK,
			responseBody:  `[{"id":123,"trackName":}]`,
			expectError:   true,
			errorContains: "failed to parse response",
		},
		{
			name:          "Server Error - 500 Status",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `Internal error`,
			expectError:   true,
			errorContains: "LRCLib server error", // Match your struct exactly
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(server.URL, 5*time.Second)
			lyrics, err := client.GetLyrics(context.Background(), "Test", "Artist")

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectSentinel != nil {
					assert.ErrorIs(t, err, tt.expectSentinel)
				}
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, lyrics)
				assert.Equal(t, 123, lyrics.TrackID)
			}
		})
	}
}

func TestGetLyrics_RequestValidation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This ensures your client is building the URL correctly
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/search", r.URL.Path)
		assert.Equal(t, "Señorita", r.URL.Query().Get("track_name"))
		assert.Equal(t, "Shawn Mendes", r.URL.Query().Get("artist_name"))
		assert.Contains(t, r.Header.Get("User-Agent"), "lyrics-analyzer")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"trackName":"Test","artistName":"Test"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	_, err := client.GetLyrics(context.Background(), "Señorita", "Shawn Mendes")
	assert.NoError(t, err)
}

func TestGetLyrics_Context(t *testing.T) {
	t.Parallel()

	// Use a mock server that hangs to test timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)

	t.Run("Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_, err := client.GetLyrics(ctx, "Test", "Artist")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}
