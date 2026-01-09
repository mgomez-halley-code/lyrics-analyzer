package client

import "context"

// LyricsClient defines the interface for fetching lyrics
type LyricsClient interface {
	GetLyrics(ctx context.Context, track, artist string) (*LyricsData, error)
}
