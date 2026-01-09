# Lyrics Client Package

A modular Go client for fetching synced and plain lyrics from multiple APIs. This package utilizes the **Decorator Pattern** to provide a resilient, testable, and extensible architecture.

## Architecture

```
internal/client/
├── interface.go       # Shared LyricsClient interface & LyricsData type
├── retry.go           # Retry decorator (works with any client)
├── lrclib/
│   ├── client.go      # LRCLib API client implementation
│   └── client_test.go
└── [future APIs]/     # Add new API clients here
```

## Quick Start

```go
import (
    "your-project/internal/client"
    "your-project/internal/client/lrclib"
)

// 1. Initialize LRCLib client
base := lrclib.NewClient("https://lrclib.net", 10*time.Second)

// 2. Wrap with Resilience (Retry logic with Exponential Backoff + Jitter)
lyricClient := client.NewRetryDecorator(base, client.DefaultRetryConfig())

// 3. Fetch
lyrics, err := lyricClient.GetLyrics(ctx, "Never Gonna Give You Up", "Rick Astley")
if errors.Is(err, lrclib.ErrLyricsNotFound) {
    // Handle 404/Empty results gracefully
}
```

## Adding a New API Client

1. Create a new subpackage: `internal/client/newapi/`
2. Implement the `client.LyricsClient` interface
3. Transform API responses to `client.LyricsData`
4. Add tests in `newapi/client_test.go`

# Run all tests
```bash
go test -v ./internal/client/...
```

# Run unit tests only (skips live integration checks)
```bash
go test -short ./internal/client/...
```