package service

import (
	"context"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// LyricsProvider defines the interface that external clients must implement
// to provide lyrics data to the service layer.
type LyricsProvider interface {
	GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error)
}
