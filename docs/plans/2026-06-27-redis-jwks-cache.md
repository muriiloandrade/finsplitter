# Redis JWKS Cache — Implementation Plan

> **For Claude:** Use the Development Agent workflow to implement this plan task-by-task.

**Goal:** Swap in-memory JWKS cache for Valkey-backed Redis with OTel instrumentation.

**Architecture:** Valkey 9.1 in compose.infra.yml → go-redis v9 in `pkg/cache` with redisotel tracing/metrics → middleware reads JWKS from Redis with TTL → graceful fallback on Redis errors.

**Tech Stack:** Valkey 9.1, go-redis v9, redisotel, OTel 1.44

---

### Task 1: Add Valkey to compose.infra.yml

**Files:**
- Modify: `compose.infra.yml`

Add Valkey service alongside PostgreSQL with healthcheck and named volume.

### Task 2: Redis config + .env.example

**Files:**
- Modify: `internal/config/config.go`
- Modify: `.env.example`

### Task 3: go-redis dependencies

**Run:** `GOEXPERIMENT=jsonv2 go get github.com/redis/go-redis/v9 github.com/redis/go-redis/extra/redisotel/v9`

### Task 4: pkg/cache/client.go

**Files:**
- Create: `pkg/cache/client.go`

Thin go-redis wrapper with `SetJSON`, `GetJSON`, `Ping`, `Close`. OTel via `redisotel.InstrumentTracing` + `InstrumentMetrics`.

### Task 5: Middleware Redis integration

**Files:**
- Modify: `internal/gateways/http/v1/auth/middleware.go`

Replace `jwkSetMu`/`jwkSet`/`jwkSetTime` with `cache *cache.Client`. `getJWKS` checks Redis first, falls back to direct fetch. Cache logs re-enabled.

### Task 6: Wire Redis client in main.go

**Files:**
- Modify: `cmd/api/main.go`

Create Redis client, inject into auth middleware, close on shutdown.

### Task 7: Build, lint, verify

- `make code-check`
- `docker compose build api`
- `make start-infra` + verify Valkey starts
- Verify cache logs (miss → hit)

### Task 8: Commit
