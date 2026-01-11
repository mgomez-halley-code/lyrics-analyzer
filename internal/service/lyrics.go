package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// LyricsService orchestrates lyrics analysis
type LyricsService struct {
	lyricsProvider LyricsProvider
	parser         *Parser
	chorusDetector *ChorusDetector
}

// NewLyricsService creates a new lyrics service
func NewLyricsService(
	lyricsProvider LyricsProvider,
	parser *Parser,
	chorusDetector *ChorusDetector,
) *LyricsService {
	return &LyricsService{
		lyricsProvider: lyricsProvider,
		parser:         parser,
		chorusDetector: chorusDetector,
	}
}

// AnalyzeSong performs complete song analysis
func (ls *LyricsService) AnalyzeSong(ctx context.Context, track, artist string) (*model.SongAnalysisResponse, error) {
	startTime := time.Now()

	// Fetch lyrics from LRCLib API
	lyricsData, err := ls.lyricsProvider.GetLyrics(ctx, track, artist)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}

	// Build track info
	trackInfo := model.Track{
		ID:           lyricsData.TrackID,
		Name:         lyricsData.TrackName,
		Artist:       lyricsData.ArtistName,
		Album:        lyricsData.AlbumName,
		Duration:     lyricsData.Duration,
		Instrumental: lyricsData.Instrumental,
	}

	// If instrumental, return early (no lyrics to analyze)
	if lyricsData.Instrumental {
		processingTime := time.Since(startTime).Milliseconds()
		return &model.SongAnalysisResponse{
			Track: trackInfo,
			Metadata: model.Metadata{
				Source:           model.SourceLRCLib,
				Cached:           false,
				ProcessingTimeMs: processingTime,
				Timestamp:        time.Now(),
				Message:          "Instrumental track - no lyrics available",
			},
		}, nil
	}

	// Parse lyrics (prefer synced over plain)
	lines, lyricsType, hasTimestamps, err := ls.parseLyrics(lyricsData)
	if err != nil {
		return nil, err
	}

	// No lyrics available
	if lines == nil {
		processingTime := time.Since(startTime).Milliseconds()
		return &model.SongAnalysisResponse{
			Track: trackInfo,
			Metadata: model.Metadata{
				Source:           model.SourceLRCLib,
				Cached:           false,
				ProcessingTimeMs: processingTime,
				Timestamp:        time.Now(),
				Message:          "No lyrics available for this track",
			},
		}, nil
	}

	// Build lyrics data
	lyricsInfo := &model.LyricsData{
		Type:          lyricsType,
		HasTimestamps: hasTimestamps,
		TotalLines:    len(lines),
		Lines:         lines,
	}

	// Detect chorus (graceful degradation - don't fail if chorus detection fails)
	var chorus *model.Chorus
	if ls.chorusDetector != nil {
		chorus = ls.chorusDetector.DetectChorus(lines)
	}

	// If chorus detection failed or is nil, return empty chorus result
	if chorus == nil {
		chorus = &model.Chorus{Detected: false}
	}

	structure := &model.Structure{
		Chorus: chorus,
	}

	// Calculate processing time
	processingTime := time.Since(startTime).Milliseconds()

	// Build complete response
	response := &model.SongAnalysisResponse{
		Track:     trackInfo,
		Lyrics:    lyricsInfo,
		Structure: structure,
		Metadata: model.Metadata{
			Source:           model.SourceLRCLib,
			Cached:           false,
			ProcessingTimeMs: processingTime,
			Timestamp:        time.Now(),
		},
	}

	return response, nil
}

// parseLyrics handles lyrics parsing logic, returning lines, type, and timestamp flag
func (ls *LyricsService) parseLyrics(lyricsData *model.LyricsSourceData) ([]model.LyricLine, string, bool, error) {
	// Prefer synced lyrics over plain
	if lyricsData.SyncedLyrics != "" {
		lines, err := ls.parser.ParseSyncedLyrics(lyricsData.SyncedLyrics)
		if err != nil {
			return nil, "", false, fmt.Errorf("failed to parse synced lyrics: %w", err)
		}
		return lines, model.LyricsTypeSynced, true, nil
	}

	if lyricsData.PlainLyrics != "" {
		lines, err := ls.parser.ParsePlainLyrics(lyricsData.PlainLyrics)
		if err != nil {
			return nil, "", false, fmt.Errorf("failed to parse plain lyrics: %w", err)
		}
		return lines, model.LyricsTypePlain, false, nil
	}

	// No lyrics available
	return nil, "", false, nil
}
