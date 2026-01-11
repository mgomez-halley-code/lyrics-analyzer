package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/handler"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/service"
)

// NewRouter builds the application's HTTP router and registers routes
func NewRouter(svc *service.LyricsService) http.Handler {
	r := mux.NewRouter()

	songHandler := handler.NewSongHandler(svc)
	healthHandler := handler.NewHealthHandler("1.0.0")

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/song/analyze", songHandler.Analyze).Methods(http.MethodGet)

	// Health check endpoint
	r.HandleFunc("/health", healthHandler.Handle).Methods(http.MethodGet)

	r.Use(loggingMiddleware)

	return r
}

// NewServer creates an HTTP server with sane defaults
func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Handler:      handler,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

// loggingMiddleware is a simple request logger
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
