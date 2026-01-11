package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/cache"
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// CachingProvider wraps a LyricsProvider with caching capabilities.
type CachingProvider struct {
	provider LyricsProvider
	cache    cache.Cache
	ttl      time.Duration
}

// NewCachingProvider creates a new caching provider.
func NewCachingProvider(provider LyricsProvider, cache cache.Cache, ttl time.Duration) *CachingProvider {
	return &CachingProvider{
		provider: provider,
		cache:    cache,
		ttl:      ttl,
	}
}

// GetLyrics fetches lyrics from cache or provider.
func (cp *CachingProvider) GetLyrics(ctx context.Context, track, artist string) (*model.LyricsSourceData, error) {
	cacheKey := cp.generateCacheKey(track, artist)

	// Try to get from cache
	cached, err := cp.cache.Get(ctx, cacheKey)
	if err != nil {
		// Gracefully handle cache errors - don't fail the request
		// Fall through to fetch from provider
	}

	// Cache hit - deserialize and return
	if cached != nil {
		var data model.LyricsSourceData
		if err := json.Unmarshal(cached, &data); err == nil {
			// mark as cached so callers can surface this information
			data.Cached = true
			return &data, nil
		}
		// If unmarshal fails, fall through to fetch from provider
	}

	// Cache miss - fetch from provider
	data, err := cp.provider.GetLyrics(ctx, track, artist)
	if err != nil {
		return nil, err
	}

	// ensure provider result is marked as not cached
	if data != nil {
		data.Cached = false
	}

	// Store in cache (fire and forget - don't block on cache errors)
	go func() {
		serialized, err := json.Marshal(data)
		if err != nil {
			return
		}
		// Use background context since request context may be cancelled
		_ = cp.cache.Set(context.Background(), cacheKey, serialized, cp.ttl)
	}()

	return data, nil
}

// generateCacheKey creates a deterministic cache key from track and artist.
// Format: lyrics:<hash of "artist:<artist>|track:<track>">
func (cp *CachingProvider) generateCacheKey(track, artist string) string {
	// Normalize to lowercase and trim spaces with descriptive labels
	normalized := fmt.Sprintf("artist:%s|track:%s",
		strings.ToLower(strings.TrimSpace(artist)),
		strings.ToLower(strings.TrimSpace(track)),
	)

	// Hash to keep keys short and avoid special characters
	hash := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("lyrics:%x", hash[:16]) // Use first 16 bytes (32 hex chars)
}
