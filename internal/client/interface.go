// Package client provides interfaces and implementations for fetching lyrics from external APIs.
// This package defines the LyricsClient interface that all API clients must implement,
// allowing the service layer to work with any lyrics provider without knowing implementation details.
package client

import (
	"context"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// LyricsClient defines the interface for fetching raw lyrics from external providers.
type LyricsClient interface {
	GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error)
}
