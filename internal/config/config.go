package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration values
type Config struct {
	ServerAddr      string
	LRCLibBaseURL   string
	LRCLibTimeout   time.Duration
	RetryMaxRetries int
	RetryBackoff    time.Duration
	RetryMaxBackoff time.Duration
	RetryMultiplier float64
}

// Load reads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	cfg := &Config{
		ServerAddr:      getEnv("SERVER_ADDR", ":8080"),
		LRCLibBaseURL:   getEnv("LRCLIB_BASE_URL", "https://lrclib.net"),
		LRCLibTimeout:   parseDurationOrDefault(getEnv("LRCLIB_TIMEOUT", "10s"), 10*time.Second),
		RetryMaxRetries: int(parseIntOrDefault(getEnv("RETRY_MAX_RETRIES", "3"), 3)),
		RetryBackoff:    parseDurationOrDefault(getEnv("RETRY_BACKOFF", "100ms"), 100*time.Millisecond),
		RetryMaxBackoff: parseDurationOrDefault(getEnv("RETRY_MAX_BACKOFF", "5s"), 5*time.Second),
		RetryMultiplier: parseFloatOrDefault(getEnv("RETRY_MULTIPLIER", "2.0"), 2.0),
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func parseDurationOrDefault(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return def
}

func parseIntOrDefault(s string, def int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return def
}

func parseFloatOrDefault(s string, def float64) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return def
}
