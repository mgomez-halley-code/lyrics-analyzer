package handler

import (
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

	// Use type assertions for structured error handling
	var notFoundErr *NotFoundError
	var rateLimitErr *RateLimitError
	var timeoutErr *TimeoutError

	switch {
	case errors.As(err, &notFoundErr):
		statusCode = http.StatusNotFound
		code = "not_found"
		message = "Song not found"
	case errors.As(err, &rateLimitErr):
		statusCode = http.StatusTooManyRequests
		code = "rate_limited"
		message = "Too many requests to lyrics provider"
	case errors.As(err, &timeoutErr):
		statusCode = http.StatusGatewayTimeout
		code = "timeout"
		message = "The request timed out"
	default:
		// Fallback to string matching for backwards compatibility with existing errors
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "no results found") || strings.Contains(errStr, "not found"):
			statusCode = http.StatusNotFound
			code = "not_found"
			message = "Song not found"
		case strings.Contains(errStr, "rate limit"):
			statusCode = http.StatusTooManyRequests
			code = "rate_limited"
			message = "Too many requests to lyrics provider"
		case strings.Contains(errStr, "context deadline exceeded") || strings.Contains(errStr, "timeout"):
			statusCode = http.StatusGatewayTimeout
			code = "timeout"
			message = "The request timed out"
		}
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
