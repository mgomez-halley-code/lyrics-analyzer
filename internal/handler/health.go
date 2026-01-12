package handler

import (
	"net/http"
	"time"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// HealthHandler exposes a basic health check endpoint.
//
// This handler is typically used by load balancers, orchestration platforms,
// and monitoring systems to verify that the service is running and responsive.
type HealthHandler struct {
	BaseHandler
	version string
}

// NewHealthHandler creates a new HealthHandler that reports the given
// application version in health check responses.
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{version: version}
}

// Handle responds with the current health status of the service.
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	h.RespondJSON(w, http.StatusOK, model.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   h.version,
	})
}
