# Lyrics Client Package

Resilient HTTP client for fetching lyrics from LRCLib API with retry logic and exponential backoff.

## Architecture

```
internal/client/
├── retry.go           # Retry decorator with exponential backoff
└── lrclib/
    ├── client.go      # LRCLib API client implementation
    └── client_test.go # Unit tests with table-driven approach
```

**Design Pattern:** Decorator Pattern - `RetryDecorator` wraps any `service.LyricsProvider` implementation.

## Components

### LRCLib Client (`lrclib/client.go`)

Implements `service.LyricsProvider` interface to fetch lyrics from LRCLib API.

**Features:**
- Searches lyrics by track and artist
- Handles 404 (not found) gracefully
- Validates and transforms API responses
- Returns structured `model.LyricsSourceData`

### Retry Decorator (`retry.go`)

Wraps any lyrics provider with resilience features:
- Exponential backoff with jitter
- Configurable retry attempts
- Only retries on transient errors (5xx, network issues)
- Skips retry on 404 (not found)

## Usage

```go
import (
    "github.com/mgomez-halley-code/lyrics-analyzer.git/internal/client"
    "github.com/mgomez-halley-code/lyrics-analyzer.git/internal/client/lrclib"
    "github.com/mgomez-halley-code/lyrics-analyzer.git/internal/service"
)

// 1. Create base LRCLib client
baseClient := lrclib.NewClient("https://lrclib.net", 10*time.Second)

// 2. Wrap with retry decorator for resilience
resilientClient := client.NewRetryDecorator(
    baseClient,
    client.RetryConfig{
        MaxRetries: 3,
        Backoff:    100 * time.Millisecond,
        MaxBackoff: 5 * time.Second,
        Multiplier: 2.0,
    },
)

// 3. Use in service layer
lyricsService := service.NewLyricsService(resilientClient, nil, nil)
```

## Testing

```bash
# Run all client tests
go test -v ./internal/client/...

# Run with coverage
go test -v -cover ./internal/client/...
```