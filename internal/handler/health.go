package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	version string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		version: version,
	}
}

// Handle processes health check requests
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed", nil)
		return
	}

	response := model.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   h.version,
	}

	h.respondJSON(w, http.StatusOK, response)
}

// respondJSON sends a JSON response
func (h *HealthHandler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but can't change status code at this point
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondError sends an error response
func (h *HealthHandler) respondError(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	errorResponse := model.ErrorResponse{
		Error: model.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	h.respondJSON(w, statusCode, errorResponse)
}
