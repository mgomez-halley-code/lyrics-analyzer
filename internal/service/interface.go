package service

import "github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"

// LyricsParser defines the interface for parsing lyrics
type LyricsParser interface {
	// ParseSyncedLyrics parses synced lyrics with timestamps into structured format
	ParseSyncedLyrics(syncedLyrics string) ([]model.LyricLine, error)

	// ParsePlainLyrics parses plain lyrics without timestamps
	ParsePlainLyrics(plainLyrics string) ([]model.LyricLine, error)
}
