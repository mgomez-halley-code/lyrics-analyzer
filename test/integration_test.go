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
)

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
	// Build service with mock client
	mockClient := &mockLyricsClient{}
	parser := service.NewParser()
	chorusDetector := service.NewChorusDetector()
	svc := service.NewLyricsService(mockClient, parser, chorusDetector)

	// Build router and test server
	router := server.NewRouter(svc)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/song/analyze?track=MyTrack&artist=MyArtist", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result model.SongAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Track.Name != "MyTrack" {
		t.Fatalf("unexpected track name: %s", result.Track.Name)
	}

	if result.Lyrics == nil || result.Lyrics.TotalLines == 0 {
		t.Fatalf("expected parsed lyrics in response")
	}

	if result.Metadata.Source != model.SourceLRCLib {
		t.Fatalf("unexpected metadata source: %s", result.Metadata.Source)
	}
}

func TestIntegration_Instrumental(t *testing.T) {
	mock := &mockInstrumentalClient{}
	parser := service.NewParser()
	chorusDetector := service.NewChorusDetector()
	svc := service.NewLyricsService(mock, parser, chorusDetector)

	router := server.NewRouter(svc)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/song/analyze?track=Inst&artist=NoLyrics", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result model.SongAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if result.Lyrics != nil {
		t.Fatalf("expected no lyrics for instrumental track")
	}
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
	mock := &mockErrorClient{}
	parser := service.NewParser()
	chorusDetector := service.NewChorusDetector()
	svc := service.NewLyricsService(mock, parser, chorusDetector)

	router := server.NewRouter(svc)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/song/analyze?track=Err&artist=Provider", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected non-200 status for upstream error")
	}
}

type mockErrorClient struct{}

func (m *mockErrorClient) GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error) {
	return nil, fmt.Errorf("upstream failure")
}
