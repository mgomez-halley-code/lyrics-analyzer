package lrclib

import (
	"errors"
	"fmt"
)

// Sentinel errors for specific cases
var (
	ErrLyricsNotFound = errors.New("no lyrics found for the given criteria")
)

// SearchResponse represents the raw response from LRCLib search API
// Based on official docs: https://lrclib.net/docs
type SearchResponse struct {
	ID           int     `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	SyncedLyrics string  `json:"syncedLyrics"`
	PlainLyrics  string  `json:"plainLyrics"`
}

// APIError represents an error from the LRCLib API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// ShouldRetry implements client.RetryableError interface
// Returns true for server errors (5xx), false for client errors (4xx)
func (e *APIError) ShouldRetry() bool {
	return e.StatusCode >= 500
}
