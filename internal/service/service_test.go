package service

import (
	"context"
	"errors"
	"testing"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLyricsClient is a mock implementation of client.LyricsClient
type MockLyricsClient struct {
	mock.Mock
}

func (m *MockLyricsClient) GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error) {
	args := m.Called(ctx, track, artist)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.LyricsSourceData), args.Error(1)
}

func TestLyricsService_ParseLyrics(t *testing.T) {
	tests := []struct {
		name              string
		syncedLyrics      string
		plainLyrics       string
		expectedType      string
		expectedTimestamp bool
		expectedLines     int
		shouldBeNil       bool
		expectedFirstLine string
	}{
		{
			name:              "synced lyrics",
			syncedLyrics:      "[00:10.00] First line\n[00:15.50] Second line",
			plainLyrics:       "",
			expectedType:      model.LyricsTypeSynced,
			expectedTimestamp: true,
			expectedLines:     2,
			shouldBeNil:       false,
		},
		{
			name:              "plain lyrics",
			syncedLyrics:      "",
			plainLyrics:       "First line\nSecond line",
			expectedType:      model.LyricsTypePlain,
			expectedTimestamp: false,
			expectedLines:     2,
			shouldBeNil:       false,
		},
		{
			name:              "prefer synced over plain",
			syncedLyrics:      "[00:10.00] Synced line",
			plainLyrics:       "Plain line",
			expectedType:      model.LyricsTypeSynced,
			expectedTimestamp: true,
			expectedLines:     1,
			shouldBeNil:       false,
			expectedFirstLine: "Synced line",
		},
		{
			name:              "no lyrics",
			syncedLyrics:      "",
			plainLyrics:       "",
			expectedType:      "",
			expectedTimestamp: false,
			expectedLines:     0,
			shouldBeNil:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			service := &LyricsService{
				parser: parser,
			}

			lyricsData := &model.LyricsSourceData{
				SyncedLyrics: tt.syncedLyrics,
				PlainLyrics:  tt.plainLyrics,
			}

			lines, lyricsType, hasTimestamps, err := service.parseLyrics(lyricsData)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedType, lyricsType)
			assert.Equal(t, tt.expectedTimestamp, hasTimestamps)

			if tt.shouldBeNil {
				assert.Nil(t, lines)
			} else {
				assert.NotNil(t, lines)
				assert.Len(t, lines, tt.expectedLines)

				if tt.expectedFirstLine != "" {
					assert.Equal(t, tt.expectedFirstLine, lines[0].Text)
				}
			}
		})
	}
}

func TestLyricsService_AnalyzeSong(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockLyricsClient)
		parser := NewParser()
		chorusDetector := NewChorusDetector()

		service := NewLyricsService(mockClient, parser, chorusDetector)

		ctx := context.Background()
		lyricsData := &model.LyricsSourceData{
			TrackID:      123,
			TrackName:    "Test Song",
			ArtistName:   "Test Artist",
			AlbumName:    "Test Album",
			Duration:     180,
			Instrumental: false,
			SyncedLyrics: "[00:10.00] Test line\n[00:15.00] Test line",
			PlainLyrics:  "",
		}

		mockClient.On("GetLyrics", ctx, "Test Song", "Test Artist").Return(lyricsData, nil)

		response, err := service.AnalyzeSong(ctx, "Test Song", "Test Artist")

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Test Song", response.Track.Name)
		assert.Equal(t, "Test Artist", response.Track.Artist)
		assert.Equal(t, model.LyricsTypeSynced, response.Lyrics.Type)
		assert.True(t, response.Lyrics.HasTimestamps)
		assert.Equal(t, 2, response.Lyrics.TotalLines)
		assert.Equal(t, model.SourceLRCLib, response.Metadata.Source)
		assert.GreaterOrEqual(t, response.Metadata.ProcessingTimeMs, int64(0))
		mockClient.AssertExpectations(t)
	})

	t.Run("instrumental track", func(t *testing.T) {
		mockClient := new(MockLyricsClient)
		parser := NewParser()
		chorusDetector := NewChorusDetector()

		service := NewLyricsService(mockClient, parser, chorusDetector)

		ctx := context.Background()
		lyricsData := &model.LyricsSourceData{
			TrackID:      123,
			TrackName:    "Instrumental Track",
			ArtistName:   "Test Artist",
			AlbumName:    "Test Album",
			Duration:     180,
			Instrumental: true,
			SyncedLyrics: "",
			PlainLyrics:  "",
		}

		mockClient.On("GetLyrics", ctx, "Instrumental Track", "Test Artist").Return(lyricsData, nil)

		response, err := service.AnalyzeSong(ctx, "Instrumental Track", "Test Artist")

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Instrumental Track", response.Track.Name)
		assert.True(t, response.Track.Instrumental)
		assert.Nil(t, response.Lyrics)
		assert.Equal(t, "Instrumental track - no lyrics available", response.Metadata.Message)
		mockClient.AssertExpectations(t)
	})

	t.Run("no lyrics available", func(t *testing.T) {
		mockClient := new(MockLyricsClient)
		parser := NewParser()
		chorusDetector := NewChorusDetector()

		service := NewLyricsService(mockClient, parser, chorusDetector)

		ctx := context.Background()
		lyricsData := &model.LyricsSourceData{
			TrackID:      123,
			TrackName:    "No Lyrics Track",
			ArtistName:   "Test Artist",
			AlbumName:    "Test Album",
			Duration:     180,
			Instrumental: false,
			SyncedLyrics: "",
			PlainLyrics:  "",
		}

		mockClient.On("GetLyrics", ctx, "No Lyrics Track", "Test Artist").Return(lyricsData, nil)

		response, err := service.AnalyzeSong(ctx, "No Lyrics Track", "Test Artist")

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Nil(t, response.Lyrics)
		assert.Equal(t, "No lyrics available for this track", response.Metadata.Message)
		mockClient.AssertExpectations(t)
	})

	t.Run("client error", func(t *testing.T) {
		mockClient := new(MockLyricsClient)
		parser := NewParser()
		chorusDetector := NewChorusDetector()

		service := NewLyricsService(mockClient, parser, chorusDetector)

		ctx := context.Background()
		expectedError := errors.New("client error")

		mockClient.On("GetLyrics", ctx, "Test Song", "Test Artist").Return(nil, expectedError)

		response, err := service.AnalyzeSong(ctx, "Test Song", "Test Artist")

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to fetch lyrics")
		mockClient.AssertExpectations(t)
	})

	t.Run("with chorus detection", func(t *testing.T) {
		mockClient := new(MockLyricsClient)
		parser := NewParser()
		chorusDetector := NewChorusDetector()

		service := NewLyricsService(mockClient, parser, chorusDetector)

		ctx := context.Background()
		lyricsData := &model.LyricsSourceData{
			TrackID:      123,
			TrackName:    "Song with Chorus",
			ArtistName:   "Test Artist",
			AlbumName:    "Test Album",
			Duration:     180,
			Instrumental: false,
			PlainLyrics:  "Verse line\nChorus line\nAnother verse\nChorus line\nMore verse\nChorus line",
			SyncedLyrics: "",
		}

		mockClient.On("GetLyrics", ctx, "Song with Chorus", "Test Artist").Return(lyricsData, nil)

		response, err := service.AnalyzeSong(ctx, "Song with Chorus", "Test Artist")

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Structure)
		assert.NotNil(t, response.Structure.Chorus)
		assert.True(t, response.Structure.Chorus.Detected)
		assert.Equal(t, "Chorus line", response.Structure.Chorus.Text)
		assert.Equal(t, 3, response.Structure.Chorus.Occurrences)
		mockClient.AssertExpectations(t)
	})

	t.Run("nil chorus detector - graceful degradation", func(t *testing.T) {
		mockClient := new(MockLyricsClient)
		parser := NewParser()

		// Create service without chorus detector (nil)
		service := NewLyricsService(mockClient, parser, nil)

		ctx := context.Background()
		lyricsData := &model.LyricsSourceData{
			TrackID:      123,
			TrackName:    "Test Song",
			ArtistName:   "Test Artist",
			AlbumName:    "Test Album",
			Duration:     180,
			Instrumental: false,
			PlainLyrics:  "Line one\nLine two\nLine three",
			SyncedLyrics: "",
		}

		mockClient.On("GetLyrics", ctx, "Test Song", "Test Artist").Return(lyricsData, nil)

		response, err := service.AnalyzeSong(ctx, "Test Song", "Test Artist")

		// Should succeed even without chorus detector (graceful degradation)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Lyrics)
		assert.Equal(t, 3, response.Lyrics.TotalLines)
		assert.NotNil(t, response.Structure)
		assert.NotNil(t, response.Structure.Chorus)
		assert.False(t, response.Structure.Chorus.Detected)
		mockClient.AssertExpectations(t)
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockClient := new(MockLyricsClient)
		parser := NewParser()
		chorusDetector := NewChorusDetector()

		service := NewLyricsService(mockClient, parser, chorusDetector)

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Mock client should return context.Canceled error
		mockClient.On("GetLyrics", ctx, "Test Song", "Test Artist").Return(nil, context.Canceled)

		response, err := service.AnalyzeSong(ctx, "Test Song", "Test Artist")

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to fetch lyrics")
		mockClient.AssertExpectations(t)
	})
}
