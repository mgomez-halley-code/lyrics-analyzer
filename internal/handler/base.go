package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// BaseHandler provides shared helper methods for all HTTP handlers
type BaseHandler struct{}

// RespondJSON sends a standardized JSON response
func (h *BaseHandler) RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// RespondError sends a standardized error response using the internal model
func (h *BaseHandler) RespondError(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	h.RespondJSON(w, statusCode, model.ErrorResponse{
		Error: model.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
