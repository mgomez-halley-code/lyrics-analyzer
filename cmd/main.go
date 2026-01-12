package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/cache"
	client "github.com/mgomez-halley-code/lyrics-analyzer.git/internal/client"
	lrclib "github.com/mgomez-halley-code/lyrics-analyzer.git/internal/client/lrclib"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/config"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/server"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/service"
)

func main() {
	// Load .env for local development (silently ignores missing file)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// ─────────────────────────────────────────────
	// Initialize dependencies
	// ─────────────────────────────────────────────

	// LRCLib HTTP client
	lrclibClient := lrclib.NewClient(cfg.LRCLibBaseURL, cfg.LRCLibTimeout)

	// Wrap with retry decorator
	retryCfg := client.RetryConfig{
		MaxRetries:     cfg.RetryMaxRetries,
		InitialBackoff: cfg.RetryBackoff,
		MaxBackoff:     cfg.RetryMaxBackoff,
		Multiplier:     cfg.RetryMultiplier,
	}

	retryClient := client.NewRetryDecorator(lrclibClient, retryCfg)

	// Initialize cache
	cacheInstance, err := cache.New(cache.Config{
		Type:          cfg.CacheType,
		RedisAddr:     cfg.RedisAddr,
		RedisPassword: cfg.RedisPassword,
		RedisDB:       cfg.RedisDB,
	})
	if err != nil {
		log.Fatalf("failed to initialize cache: %v", err)
	}
	defer cacheInstance.Close()

	// Wrap provider with caching layer
	cachingProvider := service.NewCachingProvider(
		retryClient,
		cacheInstance,
		cfg.CacheTTL,
	)

	// Parser and chorus detector
	parser := service.NewParser()
	chorusDetector := service.NewChorusDetector()

	// Service lifecycle context
	serviceCtx, serviceCancel := context.WithCancel(context.Background())
	defer serviceCancel()

	// Lyrics service
	svc := service.NewLyricsService(
		serviceCtx,
		cachingProvider,
		parser,
		chorusDetector,
	)

	// HTTP server
	router := server.NewRouter(svc)
	srv := server.NewServer(cfg.ServerAddr, router)

	// Channel to capture server errors
	serverErr := make(chan error, 1)

	// Start server
	go func() {
		log.Printf("starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// ─────────────────────────────────────────────
	// Graceful shutdown
	// ─────────────────────────────────────────────

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("received signal %s, shutting down", sig)

	case err := <-serverErr:
		log.Printf("server error: %v", err)
	}

	// Stop service work
	serviceCancel()

	// Allow in-flight requests to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	log.Println("server exited properly")
}
