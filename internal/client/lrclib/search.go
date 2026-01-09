package lrclib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/client"
)

// GetLyrics fetches lyrics for a track and artist
func (c *Client) GetLyrics(ctx context.Context, track, artist string) (*client.LyricsData, error) {
	// Build request
	req, err := c.buildSearchRequest(ctx, track, artist)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Validate response status
	if err := c.checkResponseStatus(resp); err != nil {
		return nil, err
	}

	// Parse response body
	results, err := c.parseSearchResponse(resp)
	if err != nil {
		return nil, err
	}

	bestResult := c.selectBestResult(results)

	return &client.LyricsData{
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

// buildSearchRequest creates an HTTP request for lyrics search
func (c *Client) buildSearchRequest(ctx context.Context, track, artist string) (*http.Request, error) {
	params := url.Values{}
	params.Add("track_name", track)
	params.Add("artist_name", artist)

	fullURL := fmt.Sprintf("%s/api/search?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "lyrics-analyzer/1.0")
	return req, nil
}

// parseSearchResponse decodes JSON response body and validates results
func (c *Client) parseSearchResponse(resp *http.Response) ([]SearchResponse, error) {
	var results []SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrLyricsNotFound
	}

	return results, nil
}

// selectBestResult chooses the best match from multiple results
// Prefers results with synced lyrics over plain-only lyrics
func (c *Client) selectBestResult(results []SearchResponse) SearchResponse {
	for _, result := range results {
		if result.SyncedLyrics != "" {
			return result
		}
	}

	return results[0]
}
