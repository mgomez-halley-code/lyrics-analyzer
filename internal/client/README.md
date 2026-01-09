# Lyrics Client Package

A modular Go client for fetching synced and plain lyrics from [LRCLib](https://lrclib.net/). This package utilizes the **Decorator Pattern** to provide a resilient, testable, and extensible architecture.

## Quick Start

```go
import "your-project/internal/client"

// 1. Initialize core client
base := client.NewLRCLibClient("[https://lrclib.net](https://lrclib.net)", 10*time.Second)

// 2. Wrap with Resilience (Retry logic with Exponential Backoff + Jitter)
lyricClient := client.NewRetryDecorator(base, client.DefaultRetryConfig())

// 3. Fetch
lyrics, err := lyricClient.GetLyrics(ctx, "Never Gonna Give You Up", "Rick Astley")
if errors.Is(err, client.ErrLyricsNotFound) {
    // Handle 404/Empty results gracefully
}
```

# Run all tests
```bash
go test -v ./internal/client/...
```

# Run unit tests only (skips live integration checks)
```bash
go test -short ./internal/client/...
```