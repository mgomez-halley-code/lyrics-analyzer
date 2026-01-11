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

**Configuration:**
- Redis: 256MB max memory with LFU eviction
- API: Exposed on port 8080
- Redis: Exposed on port 6379

## Usage

### Build the Image

```bash
# From project root
docker build -f docker/Dockerfile -t lyrics-analyzer:latest .
```

### Run Production Stack

```bash
# Start full stack (API + Redis)
docker-compose -f docker/docker-compose.prod.yml up -d --build

# View logs
docker-compose -f docker/docker-compose.prod.yml logs -f lyrics-analyzer
docker-compose -f docker/docker-compose.prod.yml logs -f redis

# Check health
curl http://localhost:8080/health

# Stop stack
docker-compose -f docker/docker-compose.prod.yml down
```

### Environment Variables

Override defaults by creating a `.env` file or using `environment` in docker-compose:

```yaml
environment:
  - CACHE_TTL=48h
  - RETRY_MAX_RETRIES=5
  - LRCLIB_TIMEOUT=15s
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
| **Resource Limits** | None | Recommended |

### Production Deployment

For production, consider:

1. **Use specific image tags** instead of `latest`
2. **Set resource limits** in docker-compose:
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '1'
         memory: 512M
   ```
3. **Use secrets** for sensitive data (Redis password)
4. **Enable TLS** for Redis connection
5. **Use external Redis** (AWS ElastiCache, Redis Cloud)
6. **Add monitoring** (Prometheus, Grafana)
7. **Use reverse proxy** (Nginx, Traefik) with HTTPS
