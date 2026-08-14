<!-- Context: pkg/cache | Priority: high | Version: 1.0 | Updated: 2026-08-12 -->

# pkg/cache — Valkey/Redis client wrapper

**Purpose**: Thin, typed wrapper around go-redis/v9 (v9.22.0) for Valkey/Redis. Provides JSON-safe get/set with TTL and OpenTelemetry instrumentation enabled by default on every client. Used for JWKS caching (15min TTL) in the auth middleware.

## Ownership

- `pkg/cache/client.go` — `Client`, `Option`, `New`, `Close`, `Ping`, `SetJSON`, `GetJSON`, `GetBytes`
- `pkg/cache/client_test.go`, `pkg/cache/cache_unit_test.go` — tests
- NOT: cache key naming/policies (consumer-owned), JWKS fetch + verify logic (`internal/gateways/http/v1/auth/middleware.go`), cache connection config (`internal/config`)
- NOT: failing a request when the cache is down — see Graceful degradation below

## Local Contracts

- Constructed only via `cache.New(ctx, url, opts...)`; `url` must be `redis://` or `rediss://` (parsed by `redis.ParseURL`). Zero value is not usable.
- OTel is mandatory: `redisotel.InstrumentTracing` + `InstrumentMetrics` run on every client; if either fails, `New` closes the client and returns an error.
- Typed API: `SetJSON(ctx, key, val, ttl)` marshals with `encoding/json`; `GetJSON(ctx, key, dest)` returns `(found bool, err)` — `(false, nil)` on miss (`redis.Nil`); `GetBytes` returns `(nil, nil)` on miss.
- TTL semantics: zero or negative TTL = no expiration.
- Errors wrapped with context: `cache parse url:`, `cache marshal:`, `cache get:`, `cache unmarshal:`.
- Callers MUST call `Close()` (releases the pool). `Ping(ctx)` checks connectivity.
- Functional options pattern: `type Option func(*Client)`.

## Work Guidance

Graceful degradation is a hard rule: a cache failure must never fail the request — consumers log a warning (`logctx.FromCtx(ctx).WarnContext`) and fall back to the source of truth. JWKS caching follows this (15min TTL, direct fetch on cache error).

```go
c, err := cache.New(ctx, cfg.CacheURL)
if err != nil { return err }
defer c.Close()

if found, err := c.GetJSON(ctx, "jwks:key", &jwks); err == nil && found {
    return jwks, nil // cache hit
}
jwks = fetchJWKS(ctx) // miss or cache down → fallback to direct fetch
if err := c.SetJSON(ctx, "jwks:key", jwks, 15*time.Minute); err != nil {
    logctx.FromCtx(ctx).WarnContext(ctx, "jwks cache write failed", slog.Any("error", err))
}
```

Add new operations as typed methods delegating to `c.raw`, wrapping errors with a short `cache <op>:` prefix. Keep the OTel wiring inside `New` — never let a caller disable it.

## 📂 Codebase References

- `pkg/cache/client.go` — client implementation
- `pkg/cache/client_test.go`, `pkg/cache/cache_unit_test.go` — tests
- `internal/gateways/http/v1/auth/middleware.go` — JWKS cache consumer (15min TTL, graceful fallback)

## Verification

```bash
GOEXPERIMENT=jsonv2 go test ./pkg/cache/...
make test
make code-check
```

## Child DOX Index

No child AGENTS.md files needed — leaf package. Root AGENTS.md owns global standards.
