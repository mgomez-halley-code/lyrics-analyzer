// Package client provides interfaces and implementations for fetching lyrics from external APIs.
// This package defines the LyricsClient interface that all API clients must implement,
// allowing the service layer to work with any lyrics provider without knowing implementation details.
package client

import "context"

// LyricsData is the shared internal representation for lyrics fetched from any API.
// All API clients transform their responses to this format, providing a consistent
// data structure regardless of the underlying API provider.
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

// LyricsClient defines the interface for fetching lyrics
// All API client implementations must satisfy this interface
type LyricsClient interface {
	GetLyrics(ctx context.Context, track, artist string) (*LyricsData, error)
}
