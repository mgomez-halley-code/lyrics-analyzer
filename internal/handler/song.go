package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/service"
)

// SongHandler handles song analysis requests
type SongHandler struct {
	lyricsService *service.LyricsService
}

// NewSongHandler creates a new song handler
func NewSongHandler(lyricsService *service.LyricsService) *SongHandler {
	return &SongHandler{
		lyricsService: lyricsService,
	}
}

// Analyze handles song analysis requests
// The router will ensure this is only called for GET requests
func (h *SongHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	track := strings.TrimSpace(r.URL.Query().Get("track"))
	artist := strings.TrimSpace(r.URL.Query().Get("artist"))

	if track == "" || artist == "" {
		h.respondError(w, http.StatusBadRequest, "missing_parameter", "Track and artist are required", nil)
		return
	}

	response, err := h.lyricsService.AnalyzeSong(r.Context(), track, artist)
	if err != nil {
		h.handleServiceError(w, track, artist, err)
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// handleServiceError maps service-layer errors to appropriate HTTP responses
func (h *SongHandler) handleServiceError(w http.ResponseWriter, track, artist string, err error) {
	statusCode := http.StatusInternalServerError
	code := "internal_error"
	message := "Failed to analyze song"

	// Type-safe error handling with errors.Is()
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
	default:
		// Use default: HTTP 500 Internal Server Error
	}

	h.respondError(w, statusCode, code, message, map[string]string{
		"track":  track,
		"artist": artist,
		"debug":  err.Error(),
	})
}

// respondJSON sends a JSON response
func (h *SongHandler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondError sends an error response
func (h *SongHandler) respondError(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	errorResponse := model.ErrorResponse{
		Error: model.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	h.respondJSON(w, statusCode, errorResponse)
}
