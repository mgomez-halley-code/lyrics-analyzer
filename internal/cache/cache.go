// Package cache provides caching interfaces and implementations for lyrics data.
// It defines a generic cache interface that can be implemented by different backends
// (in-memory, Redis, etc.) allowing the service layer to cache lyrics without
// coupling to a specific cache implementation.
package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching lyrics data.
// Implementations must be safe for concurrent use.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Close() error
}
