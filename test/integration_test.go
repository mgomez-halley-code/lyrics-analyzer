package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/server"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a test HTTP server with the given mock client
func setupTestServer(mockClient service.LyricsProvider) *httptest.Server {
	parser := service.NewParser()
	chorusDetector := service.NewChorusDetector()
	svc := service.NewLyricsService(context.Background(), mockClient, parser, chorusDetector)
	router := server.NewRouter(svc)
	return httptest.NewServer(router)
}

type mockLyricsClient struct{}

func (m *mockLyricsClient) GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error) {
	// Simple synthetic lyrics with a repeated line to exercise chorus detection
	return &model.LyricsSourceData{
		TrackID:      42,
		TrackName:    track,
		ArtistName:   artist,
		AlbumName:    "TestAlbum",
		Duration:     180,
		Instrumental: false,
		SyncedLyrics: "",
		PlainLyrics:  "Hello world\nThis is a test\nHello world",
	}, nil
}

func TestIntegration_SongAnalyze(t *testing.T) {
	ts := setupTestServer(&mockLyricsClient{})
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/song/analyze?track=MyTrack&artist=MyArtist", nil)
	require.NoError(t, err, "failed to create request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "request failed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result model.SongAnalysisResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "failed to decode response")

	assert.Equal(t, "MyTrack", result.Track.Name)
	assert.NotNil(t, result.Lyrics, "expected lyrics in response")
	assert.Greater(t, result.Lyrics.TotalLines, 0, "expected parsed lyrics lines")
	assert.Equal(t, model.SourceLRCLib, result.Metadata.Source)
}

func TestIntegration_Instrumental(t *testing.T) {
	ts := setupTestServer(&mockInstrumentalClient{})
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/song/analyze?track=Inst&artist=NoLyrics", nil)
	require.NoError(t, err, "failed to create request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "request failed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result model.SongAnalysisResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "failed to decode response")

	assert.Nil(t, result.Lyrics, "expected no lyrics for instrumental track")
}

// mockInstrumentalClient returns an instrumental result
type mockInstrumentalClient struct{}

func (m *mockInstrumentalClient) GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error) {
	return &model.LyricsSourceData{
		TrackID:      1,
		TrackName:    track,
		ArtistName:   artist,
		Instrumental: true,
	}, nil
}

func TestIntegration_UpstreamError(t *testing.T) {
	ts := setupTestServer(&mockErrorClient{})
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/song/analyze?track=Err&artist=Provider", nil)
	require.NoError(t, err, "failed to create request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "request failed")
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "expected non-200 status for upstream error")
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

type mockErrorClient struct{}

func (m *mockErrorClient) GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error) {
	return nil, fmt.Errorf("upstream failure")
}
