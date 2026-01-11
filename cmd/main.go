package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	client "github.com/mgomez-halley-code/lyrics-analyzer.git/internal/client"
	lrclib "github.com/mgomez-halley-code/lyrics-analyzer.git/internal/client/lrclib"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/config"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/server"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env for local development (silently ignores missing file)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// 1. Initialize dependencies
	// LRCLib HTTP client
	rawClient := lrclib.NewClient(cfg.LRCLibBaseURL, cfg.LRCLibTimeout)

	// Wrap with retry decorator (use config values)
	retryCfg := client.RetryConfig{
		MaxRetries:     cfg.RetryMaxRetries,
		InitialBackoff: cfg.RetryBackoff,
		MaxBackoff:     cfg.RetryMaxBackoff,
		Multiplier:     cfg.RetryMultiplier,
	}

	retryClient := client.NewRetryDecorator(rawClient, retryCfg)

	// Parser and chorus detector used by the service
	parser := service.NewParser()
	chorusDetector := service.NewChorusDetector()

	// Service
	svc := service.NewLyricsService(retryClient, parser, chorusDetector)

	// Build router and server
	r := server.NewRouter(svc)
	srv := server.NewServer(cfg.ServerAddr, r)

	// Start server in background
	go func() {
		log.Printf("starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited properly")
}
