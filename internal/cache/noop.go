package cache

import (
	"context"
	"time"
)

// NoOpCache is a cache that does nothing.
// Useful for disabling caching or testing.
type NoOpCache struct{}

// NewNoOpCache creates a new no-op cache.
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

// Get always returns nil (cache miss).
func (c *NoOpCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

// Set does nothing.
func (c *NoOpCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

// Delete does nothing.
func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

// Close does nothing.
func (c *NoOpCache) Close() error {
	return nil
}
