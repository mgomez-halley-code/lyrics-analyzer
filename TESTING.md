# Testing Guide - Lyrics Analyzer API

## Quick Start

### 1. Setup Environment (Optional)

```bash
# Copy example config (optional - has sensible defaults)
cp .env.example .env

# Edit .env if needed (e.g., set REDIS_PASSWORD for local dev)
```

### 2. Start Redis (Optional for Caching)

```bash
# Development: Start Redis
docker-compose -f docker/docker-compose.dev.yml up -d
```

**Note:** You can run without Redis using `CACHE_TYPE=none` (see below).

### 3. Start the Server

```bash
# Method 1: Run with Redis cache (default)
go run ./cmd/main.go

# Method 2: Run without cache
CACHE_TYPE=none go run ./cmd/main.go

# Method 3: Build and run
go build -o lyrics-analyzer ./cmd
./lyrics-analyzer
```

The server will start on `http://localhost:8080` by default.


## Available Endpoints

### 1. Health Check Endpoint

**Endpoint:** `GET /health`

**Purpose:** Check if the service is running (for load balancers, Kubernetes probes, etc.)

**Example Request:**
```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "ok",
  "timestamp": "2026-01-10T17:29:51.384Z",
  "version": "1.0.0"
}
```

**Status Code:** `200 OK`


### 2. Song Analysis Endpoint

**Endpoint:** `GET /api/song/analyze`

**Purpose:** Analyze lyrics for a song by track name and artist

**Query Parameters:**
- `track` (required): Song/track name
- `artist` (required): Artist name


## Testing Examples

### Example 1: Analyze "Bohemian Rhapsody" by Queen

```bash
curl "http://localhost:8080/api/song/analyze?track=Bohemian%20Rhapsody&artist=Queen"
```

**Expected Response:** Full lyrics with timestamps, chorus detection, and metadata


### Example 2: Analyze "Imagine" by John Lennon

```bash
curl "http://localhost:8080/api/song/analyze?track=Imagine&artist=John%20Lennon"
```


### Example 3: Test Missing Parameters (Error Case)

```bash
curl "http://localhost:8080/api/song/analyze?track=Hello"
```

**Expected Response:**
```json
{
  "error": {
    "code": "missing_parameter",
    "message": "Track and artist are required",
    "details": null
  }
}
```

**Status Code:** `400 Bad Request`

### Example 4: Test Song Not Found (Error Case)

```bash
curl "http://localhost:8080/api/song/analyze?track=NonExistentSong&artist=FakeArtist"
```

**Expected Response:**
```json
{
  "error": {
    "code": "not_found",
    "message": "Song not found",
    "details": {
      "track": "NonExistentSong",
      "artist": "FakeArtist",
      "debug": "..."
    }
  }
}
```

**Status Code:** `404 Not Found`

## Testing with Tools

### Using cURL

```bash
# Pretty-print JSON response (requires jq)
curl "http://localhost:8080/api/song/analyze?track=Imagine&artist=John%20Lennon" | jq

# Save response to file
curl "http://localhost:8080/api/song/analyze?track=Imagine&artist=John%20Lennon" -o response.json

# Show HTTP headers
curl -i "http://localhost:8080/health"
```

## Response Structure

### Successful Analysis Response

```json
{
  "track": {
    "id": 123456,
    "name": "Song Name",
    "artist": "Artist Name",
    "album": "Album Name",
    "duration": 240,
    "instrumental": false
  },
  "lyrics": {
    "type": "synced",
    "hasTimestamps": true,
    "totalLines": 50,
    "lines": [
      {
        "lineNumber": 1,
        "timestamp": "00:00.78",
        "seconds": 0.78,
        "text": "Lyric line text",
        "wordCount": 9
      }
    ]
  },
  "structure": {
    "chorus": {
      "detected": true,
      "text": "Chorus text",
      "occurrences": 3,
      "lineNumbers": [10, 25, 40]
    }
  },
  "metadata": {
    "source": "lrclib",
    "cached": false,
    "processingTimeMs": 1234,
    "timestamp": "2026-01-10T17:30:04.548Z"
  }
}
```

## Running Automated Tests

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run tests in a specific package
go test ./internal/service/... -v
```

### Integration Tests

```bash
# Run integration tests
go test ./test/... -v
```

## Configuration

You can configure the server using environment variables:

```bash
# Server configuration
export SERVER_ADDR=":8080"

# LRCLib API configuration
export LRCLIB_BASE_URL="https://lrclib.net"
export LRCLIB_TIMEOUT="10s"

# Retry configuration
export RETRY_MAX_RETRIES="3"
export RETRY_BACKOFF="100ms"
export RETRY_MAX_BACKOFF="5s"
export RETRY_MULTIPLIER="2"

# Start server with custom config
go run ./cmd/main.go
```

## Troubleshooting

### Server Won't Start

**Error:** `bind: address already in use`

**Solution:** Another process is using port 8080. Change the port:
```bash
export SERVER_ADDR=":9000"
go run ./cmd/main.go
```

## Health Check for Production

### Docker Health Check
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
```

### Kubernetes Liveness Probe
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

### Kubernetes Readiness Probe
```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Docker Production Testing

**Prerequisites:** Create a `.env` file with `REDIS_PASSWORD` before running production Docker stack.

```bash
# Generate secure password
echo "REDIS_PASSWORD=$(openssl rand -base64 32)" > .env
```

**Note:** The `.env` file is gitignored for security. This file is required for production deployment.

🐳 **For complete Docker deployment instructions and cache testing, see [docker/README.md](docker/README.md)**

## Sample Test Songs

Here are some songs that work well for testing:

| Track | Artist | Notes |
|-------|--------|-------|
| Bohemian Rhapsody | Queen | Long song with complex structure |
| Imagine | John Lennon | Classic with clear chorus |
| Hotel California | Eagles | Multi-verse with chorus |
| Smells Like Teen Spirit | Nirvana | Repetitive chorus detection |
| Let It Be | The Beatles | Simple structure |

## API Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `missing_parameter` | Track or artist parameter missing |
| 404 | `not_found` | Song not found in lyrics database |
| 429 | `rate_limited` | Too many requests to upstream API |
| 500 | `internal_error` | Server error |
| 504 | `timeout` | Request timed out |
