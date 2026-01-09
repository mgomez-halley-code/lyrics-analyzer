package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Sentinel errors for specific cases
var (
	ErrLyricsNotFound = errors.New("no lyrics found for the given criteria")
)

// LRCLibClient handles communication with LRCLib API
type LRCLibClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewLRCLibClient creates a new LRCLib client
func NewLRCLibClient(baseURL string, timeout time.Duration) *LRCLibClient {
	return &LRCLibClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// LRCLibResponse represents the raw response from LRCLib API
// Based on official docs: https://lrclib.net/docs
type LRCLibResponse struct {
	ID           int     `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	SyncedLyrics string  `json:"syncedLyrics"`
	PlainLyrics  string  `json:"plainLyrics"`
}

// LyricsData is our internal representation after fetching from API
type LyricsData struct {
	TrackID      int
	TrackName    string
	ArtistName   string
	AlbumName    string
	Duration     int
	Instrumental bool
	SyncedLyrics string
	PlainLyrics  string
}

// APIError represents an error from the LRCLib API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// GetLyrics fetches lyrics for a track and artist
func (c *LRCLibClient) GetLyrics(ctx context.Context, track, artist string) (*LyricsData, error) {
	// Build URL with query parameters
	apiURL := fmt.Sprintf("%s/api/search", c.baseURL)
	params := url.Values{}
	params.Add("track_name", track)
	params.Add("artist_name", artist)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "lyrics-analyzer/1.0")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-success status codes before parsing body
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, ErrLyricsNotFound

		case http.StatusInternalServerError:
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    "LRCLib server error",
			}

		case http.StatusServiceUnavailable:
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    "LRCLib service unavailable",
			}

		default:
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("Unexpected status code: %d", resp.StatusCode),
			}
		}
	}

	// Parse response body - only reached if status is 200 OK
	var results []LRCLibResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrLyricsNotFound
	}

	bestResult := c.selectBestResult(results)

	return &LyricsData{
		TrackID:      bestResult.ID,
		TrackName:    bestResult.TrackName,
		ArtistName:   bestResult.ArtistName,
		AlbumName:    bestResult.AlbumName,
		Duration:     int(bestResult.Duration),
		Instrumental: bestResult.Instrumental,
		SyncedLyrics: bestResult.SyncedLyrics,
		PlainLyrics:  bestResult.PlainLyrics,
	}, nil
}

// selectBestResult chooses the best match from multiple results
// Prefers results with synced lyrics over plain-only lyrics
func (c *LRCLibClient) selectBestResult(results []LRCLibResponse) LRCLibResponse {
	for _, result := range results {
		if result.SyncedLyrics != "" {
			return result
		}
	}

	return results[0]
}
