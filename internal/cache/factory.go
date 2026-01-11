package cache

import (
	"fmt"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/cache/redis"
)

// Config holds cache configuration
type Config struct {
	Type          string // "redis" or "none"
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

// New creates a cache based on the configuration.
// Returns a no-op cache if type is "none".
func New(cfg Config) (Cache, error) {
	switch cfg.Type {
	case "redis":
		return redis.NewCache(redis.Config{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})

	case "none":
		return NewNoOpCache(), nil

	default:
		return nil, fmt.Errorf("unknown cache type: %s (supported: redis, none)", cfg.Type)
	}
}
