package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/service"
)

// SongHandler handles HTTP requests related to song analysis.
type SongHandler struct {
	BaseHandler   // Embed base helpers for consistent responses
	lyricsService *service.LyricsService
}

// NewSongHandler creates a new SongHandler with the given LyricsService dependency.
func NewSongHandler(lyricsService *service.LyricsService) *SongHandler {
	return &SongHandler{lyricsService: lyricsService}
}

// Analyze handles requests to analyze a song's lyrics and structure.
func (h *SongHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	track := strings.TrimSpace(r.URL.Query().Get("track"))
	artist := strings.TrimSpace(r.URL.Query().Get("artist"))

	if track == "" || artist == "" {
		h.RespondError(
			w,
			http.StatusBadRequest,
			"missing_parameter",
			"Track and artist are required",
			nil,
		)
		return
	}

	response, err := h.lyricsService.AnalyzeSong(r.Context(), track, artist)
	if err != nil {
		h.handleServiceError(w, track, artist, err)
		return
	}

	h.RespondJSON(w, http.StatusOK, response)
}

// handleServiceError maps service-layer errors to HTTP responses.
func (h *SongHandler) handleServiceError(
	w http.ResponseWriter,
	track, artist string,
	err error,
) {
	statusCode := http.StatusInternalServerError
	code := "internal_error"
	message := "Failed to analyze song"

	switch {
	case errors.Is(err, model.ErrNotFound):
		statusCode = http.StatusNotFound
		code = "not_found"
		message = "Song not found"

	case errors.Is(err, model.ErrRateLimited):
		statusCode = http.StatusTooManyRequests
		code = "rate_limited"
		message = "Too many requests to lyrics provider"

	case errors.Is(err, context.DeadlineExceeded):
		statusCode = http.StatusGatewayTimeout
		code = "timeout"
		message = "The request timed out"
	}

	h.RespondError(w, statusCode, code, message, map[string]string{
		"track":  track,
		"artist": artist,
		"debug":  err.Error(),
	})
}
