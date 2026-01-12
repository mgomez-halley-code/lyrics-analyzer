package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/client/lrclib"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// LyricsService orchestrates lyrics analysis
type LyricsService struct {
	ctx            context.Context // Base context for service lifecycle
	lyricsProvider LyricsProvider
	parser         *Parser
	chorusDetector *ChorusDetector
}

// NewLyricsService creates a new lyrics service with a base context for lifecycle management
func NewLyricsService(
	ctx context.Context,
	lyricsProvider LyricsProvider,
	parser *Parser,
	chorusDetector *ChorusDetector,
) *LyricsService {
	return &LyricsService{
		ctx:            ctx,
		lyricsProvider: lyricsProvider,
		parser:         parser,
		chorusDetector: chorusDetector,
	}
}

// AnalyzeSong performs complete song analysis
func (ls *LyricsService) AnalyzeSong(
	ctx context.Context,
	track, artist string,
) (*model.SongAnalysisResponse, error) {
	startTime := time.Now()

	// Abort early if service is shutting down
	select {
	case <-ls.ctx.Done():
		return nil, fmt.Errorf("service is shutting down: %w", ls.ctx.Err())
	default:
	}

	// Fetch lyrics
	lyricsData, err := ls.lyricsProvider.GetLyrics(ctx, track, artist)
	if err != nil {
		if errors.Is(err, lrclib.ErrLyricsNotFound) {
			return nil, fmt.Errorf("%w: %v", model.ErrNotFound, err)
		}
		if errors.Is(err, lrclib.ErrRateLimited) {
			return nil, fmt.Errorf("%w: %v", model.ErrRateLimited, err)
		}
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}

	trackInfo := model.Track{
		ID:           lyricsData.TrackID,
		Name:         lyricsData.TrackName,
		Artist:       lyricsData.ArtistName,
		Album:        lyricsData.AlbumName,
		Duration:     lyricsData.Duration,
		Instrumental: lyricsData.Instrumental,
	}

	// Instrumental track → return early
	if lyricsData.Instrumental {
		return &model.SongAnalysisResponse{
			Track: trackInfo,
			Metadata: ls.buildMetadata(
				startTime,
				lyricsData.Cached,
				"Instrumental track - no lyrics available",
			),
		}, nil
	}

	// Parse lyrics
	lines, lyricsType, hasTimestamps, err := ls.parseLyrics(lyricsData)
	if err != nil {
		return nil, err
	}

	// No lyrics available
	if lines == nil {
		return &model.SongAnalysisResponse{
			Track: trackInfo,
			Metadata: ls.buildMetadata(
				startTime,
				lyricsData.Cached,
				"No lyrics available for this track",
			),
		}, nil
	}

	lyricsInfo := &model.LyricsData{
		Type:          lyricsType,
		HasTimestamps: hasTimestamps,
		TotalLines:    len(lines),
		Lines:         lines,
	}

	// Detect chorus (non-fatal)
	var chorus *model.Chorus
	if ls.chorusDetector != nil {
		chorus = ls.chorusDetector.DetectChorus(lines)
	}
	if chorus == nil {
		chorus = &model.Chorus{Detected: false}
	}

	structure := &model.Structure{
		Chorus: chorus,
	}

	return &model.SongAnalysisResponse{
		Track:     trackInfo,
		Lyrics:    lyricsInfo,
		Structure: structure,
		Metadata:  ls.buildMetadata(startTime, lyricsData.Cached, ""),
	}, nil
}

// buildMetadata centralizes metadata creation
func (ls *LyricsService) buildMetadata(
	startTime time.Time,
	cached bool,
	message string,
) model.Metadata {
	return model.Metadata{
		Source:           model.SourceLRCLib,
		Cached:           cached,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		Timestamp:        time.Now(),
		Message:          message,
	}
}

// parseLyrics handles lyrics parsing logic
func (ls *LyricsService) parseLyrics(
	lyricsData *model.LyricsSourceData,
) ([]model.LyricLine, string, bool, error) {

	// Prefer synced lyrics
	if lyricsData.SyncedLyrics != "" {
		lines, err := ls.parser.ParseSyncedLyrics(lyricsData.SyncedLyrics)
		if err != nil {
			return nil, "", false, fmt.Errorf("failed to parse synced lyrics: %w", err)
		}
		return lines, model.LyricsTypeSynced, true, nil
	}

	// Fallback to plain lyrics
	if lyricsData.PlainLyrics != "" {
		lines, err := ls.parser.ParsePlainLyrics(lyricsData.PlainLyrics)
		if err != nil {
			return nil, "", false, fmt.Errorf("failed to parse plain lyrics: %w", err)
		}
		return lines, model.LyricsTypePlain, false, nil
	}

	return nil, "", false, nil
}
