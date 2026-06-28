# Redis-Based JWKS Cache with OpenTelemetry

**Date**: 2026-06-27
**Status**: Design — Approved
**Related**: [Auth Logto Integration](2026-03-20-auth-logto-integration.md)

## Overview

Replace the in-memory `sync.RWMutex` + `jwk.Set` cache in the auth middleware
with a Valkey-backed cache. This adds:

- A Valkey 9.1 service in `compose.infra.yml`
- A `pkg/cache` Go module wrapping `go-redis/v9` with OTel instrumentation
- Redis-based JWKS storage with TTL (15 min)
- Cache hit/miss logs backed by real Redis operations
- Graceful degradation when Valkey is unavailable

## Components

### 1. Valkey Service (`compose.infra.yml`)

```yaml
valkey:
    profiles:
        - infra
    image: valkey/valkey:9.1.0-trixie@sha256:4963247afc4cd33c7d3b2d2816b9f7f8eeebab148d29056c2ca4d7cbc966f2d9
    container_name: finsplitter-valkey
    restart: unless-stopped
    healthcheck:
        test: ["CMD", "valkey-cli", "ping"]
        start_period: 5s
        interval: 10s
        retries: 5
        timeout: 3s
    ports:
        - "${VALKEY_PORT:-6379}:6379"
    networks:
        - finsplitter
    volumes:
        - valkey-data:/data

volumes:
    valkey-data:
        driver: local
```

### 2. Config (`internal/config/config.go`)

New `Redis` struct:

```go
type Redis struct {
    URL string `conf:"env:REDIS_URL,default:redis://localhost:6379/0"`
}
```

Added to `Config`:

```go
type Config struct {
    App   Application
    Env   Environment
    DB    Database
    Redis Redis          // new
    OTel  OpenTelemetry
    Logto Logto
}
```

### 3. `.env.example`

```env
# ============================================================
# Valkey Configuration
# ============================================================
REDIS_URL=redis://host.docker.internal:6379/0
VALKEY_PORT=6379
```

### 4. `pkg/cache/client.go`

Thin go-redis wrapper with functional options, exposing only typed methods:

```go
package cache

type Client struct {
    raw redis.UniversalClient
}

func New(ctx context.Context, url string, opts ...Option) (*Client, error)
func (c *Client) SetJSON(ctx context.Context, key string, val any, ttl time.Duration) error
func (c *Client) GetJSON(ctx context.Context, key string, dest any) (found bool, err error)
func (c *Client) Del(ctx context.Context, keys ...string) error
func (c *Client) Ping(ctx context.Context) error
func (c *Client) Close() error
```

OTel instrumentation via `redisotel`:

```go
import "github.com/redis/go-redis/extra/redisotel/v9"

rdb := redis.NewClient(opts)
redisotel.InstrumentTracing(rdb)  // spans per command
redisotel.InstrumentMetrics(rdb)  // counters + histograms
```

Both use the global OTel tracer/meter provider — consistent with the rest of the
application (otelhttp, otelpgx, otelchi).

### 5. Middleware Changes (`internal/gateways/http/v1/auth/middleware.go`)

`NewMiddleware` accepts an additional `*cache.Client` parameter. The
`Middleware` struct replaces `jwkSetMu`/`jwkSet`/`jwkSetTime` with a single
`cache *cache.Client` field.

`getJWKS` becomes:

```
1. Redis GET "jwks:keyset"
   → Hit: logger.Debug("JWKS cache hit"), unmarshal, return
2. Miss: logger.Debug("JWKS cache miss, fetching from Logto")
   → Fetch from Logto via jwkfetch.Client
   → Redis SET "jwks:keyset" <json> TTL=15min
   → return
3. Redis error: logger.Warn, fall back to direct fetch (graceful degradation)
```

### 6. Wiring (`cmd/api/main.go`)

```go
redisClient, err := cache.New(ctx, cfg.Redis.URL)
// ...
newAuthMiddleware(logger, cfg, userRepo, redisClient)

// shutdown:
redisClient.Close()
```

## Data Flow

```
Request with Bearer token
  → Middleware.requireAuth()
    → parseAndValidate()
      → getJWKS(ctx)
         ├── Redis GET "jwks:keyset"
         │    ├── HIT  → log "cache hit",  unmarshal, return
         │    └── MISS → log "cache miss", fetch from Logto
         │               Redis SET "jwks:keyset" <json> EX 900
         │               return
         └── Redis ERROR → log warn, fetch from Logto directly
  → jwt.Parse(tok, WithKeySet(keyset))
  → jwt.Validate(tok, issuer, audience)
```

## Dependencies

```
go get github.com/redis/go-redis/v9
go get github.com/redis/go-redis/extra/redisotel/v9
```

## Exit Criteria

- [ ] `make start-infra` starts PostgreSQL + Valkey
- [ ] `make code-check` passes (build + lint)
- [ ] First authenticated request logs "JWKS cache miss", subsequent requests log "JWKS cache hit"
- [ ] Stopping Valkey container triggers graceful fallback (warn log + direct fetch)
- [ ] Restarting Valkey re-fetches JWKS on next request (cache evicted)
- [ ] OTel spans visible for Redis GET/SET commands
