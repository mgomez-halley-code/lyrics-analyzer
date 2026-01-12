# Docker Configuration

This directory contains Docker configurations for both development and production deployments.

## Files

### `docker-compose.dev.yml`

Development compose file for local Redis instance:

**Services:**
- `redis` - Redis 7 with persistence for development

**Features:**
- Health checks enabled
- Persistent volume (data survives container restarts)
- Port 6379 exposed to host
- LFU eviction policy not configured (no memory limits in dev)

**Usage:**
```bash
# Start Redis for development
docker-compose -f docker/docker-compose.dev.yml up -d

# Stop Redis
docker-compose -f docker/docker-compose.dev.yml down
```

Run your Go application locally with `go run ./cmd/main.go` - it will connect to this Redis instance.

### `docker-compose.prod.yml`

Production stack with both the API service and Redis:

**Services:**
- `lyrics-analyzer` - Go API service
- `redis` - Redis 7 with persistence

**Features:**
- Health checks for both services
- Automatic restart policies
- Persistent Redis data
- Custom bridge network
- Environment-based configuration
- Service dependencies (wait for Redis)
- Resource limits (CPU/memory) to prevent OOM kills
- Redis password authentication
- Network isolation (Redis not exposed to host)

## Usage

### Build the Image

```bash
# From project root
docker build -f docker/Dockerfile -t lyrics-analyzer:latest .
```

### Run Production Stack

**REQUIRED:** You MUST set `REDIS_PASSWORD` before running. The stack will fail to start without it.

```bash
# 1. Create .env file with Redis password (from project root)
echo "REDIS_PASSWORD=$(openssl rand -base64 32)" > .env

# 2. Start full stack (API + Redis)
docker-compose -f docker/docker-compose.prod.yml --env-file .env up -d --build

# 3. Check health
curl http://localhost:8080/health

# 4. Stop stack
docker-compose -f docker/docker-compose.prod.yml down
```

### Test Production Stack with Redis Caching

Verify the complete production setup including cache performance:

```bash
# 1. View logs
docker-compose -f docker/docker-compose.prod.yml logs -f lyrics-analyzer
docker-compose -f docker/docker-compose.prod.yml logs -f redis

# 2. Test cache MISS (first request - fetches from LRCLib API)
curl "http://localhost:8080/api/song/analyze?track=Hotel%20California&artist=Eagles"
# Response: "cached": false, "processingTimeMs": ~2000 (2 seconds)

# 3. Test cache HIT (second request - instant from Redis)
curl "http://localhost:8080/api/song/analyze?track=Hotel%20California&artist=Eagles"
# Response: "cached": true, "processingTimeMs": 0 (instant)

# 4. Verify Redis is working (inside container)
docker exec lyrics-redis redis-cli -a "$(cat .env | cut -d= -f2)" DBSIZE
# Should show: 1 (or more if you tested multiple songs)

# 5. Check container health and resource usage
docker-compose -f docker/docker-compose.prod.yml ps
docker stats --no-stream lyrics-analyzer lyrics-redis
```

#### Inspect Redis Cache

```bash
export $(grep -v '^#' .env | xargs)

docker-compose -f docker/docker-compose.dev.yml exec redis redis-cli -a $REDIS_PASSWORD KEYS "lyrics:*"
docker-compose -f docker/docker-compose.dev.yml exec redis redis-cli -a $REDIS_PASSWORD GET "lyrics:<key>"
docker-compose -f docker/docker-compose.dev.yml exec redis redis-cli -a $REDIS_PASSWORD FLUSHDB
```

**Expected Performance:**
- **Cache MISS**: 1500-2500ms (fetches from LRCLib API)
- **Cache HIT**: 0-5ms (instant from Redis)
- **Speedup**: ~500-1000x faster on cache hits

**Security Verification:**
- ✅ Redis requires password authentication
- ✅ Redis NOT accessible from host (network isolated)
- ✅ Resource limits enforced (API: 512MB, Redis: 384MB)

### Environment Variables

Override defaults by creating a `.env` file in the project root:

```bash
# Required for production
REDIS_PASSWORD=your-strong-random-password-here

# Optional overrides
CACHE_TTL=48h
RETRY_MAX_RETRIES=5
LRCLIB_TIMEOUT=15s
```

**Generate secure password:**
```bash
openssl rand -base64 32
```

## Development vs Production

| Aspect | Development | Production |
|--------|-------------|------------|
| **Compose File** | `docker/docker-compose.dev.yml` | `docker/docker-compose.prod.yml` |
| **Services** | Redis only | API + Redis |
| **Build** | Local `go run` | Docker multi-stage |
| **Redis Data** | Persistent volume | Persistent volume |
| **Restart Policy** | `unless-stopped` | `unless-stopped` |
| **Health Checks** | Required | Required |
| **Resource Limits** | None | ✅ CPU/Memory limits |
| **Redis Password** | None | ✅ Required |
| **Redis Network** | Exposed to host | ✅ Internal only |

## Security Features

The production configuration includes:

### ✅ Implemented
1. **Resource Limits** - Prevents OOM kills and runaway containers
   - API: 1 CPU core / 512MB RAM
   - Redis: 0.5 CPU / 384MB RAM
2. **Redis Authentication** - Password-protected with `--requirepass`
3. **Network Isolation** - Redis not exposed to host, only accessible via internal network
4. **Environment Variables** - Secrets loaded from `.env` file (gitignored)
5. **Fail-Fast Security** - Container fails to start if `REDIS_PASSWORD` is not set (no weak defaults)
6. **Secure Healthchecks** - Password passed via environment variable, not visible in logs

### 🔒 Additional Hardening (Optional)

For enhanced security, consider:

1. **Use Docker Secrets** (Swarm mode) instead of environment variables
2. **Enable TLS** for Redis connection
3. **Use external managed Redis** (AWS ElastiCache, Redis Cloud)
4. **Add monitoring** (Prometheus, Grafana)
5. **Use reverse proxy** (Nginx, Traefik) with HTTPS
6. **Run as non-root user** (already implemented in Dockerfile)
7. **Enable Redis ACLs** for fine-grained permissions (Redis 6+)
