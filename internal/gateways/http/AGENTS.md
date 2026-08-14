<!-- Context: gateways/http | Priority: critical | Version: 1.0 | Updated: 2026-08-12 -->

# HTTP Gateway (global router) — chi router assembly + request ID middleware

**Purpose**: Builds the single global chi router every request flows through, and guarantees every request carries a UUID v4 request ID for tracing and logging.

## Ownership

- `global_router.go` — `NewRouter`: global middleware chain + otelchi tracing + otelchimetric server metrics
- `request_id.go` — `RequestID` middleware, `X-Request-ID` handling, UUID v4 generation
- `request_id_test.go` — middleware unit tests
- `v1/` — all routes/handlers (has its own AGENTS.md)
- NOT: handlers, use cases, domain logic, Logto clients, OpenAPI config (`api/openapi.go`)

## Local Contracts

- **Request ID**:
  - Header `X-Request-ID` (`RequestIDHeader`); absent → generate UUID v4 via gofrs/uuid.
  - Store with `logctx.WithRequestID(ctx, id)` under `logctx.RequestIDKey`; read via `logctx.GetRequestID`.
  - Custom middleware — do NOT switch to chi's `middleware.RequestID` (it generates hostname/counter IDs and exposes no generation hook; UUID v4 in every log line is a requirement).
  - `newRequestID` must never panic: UUID failure → empty string, request still served.
- **Router assembly** (`NewRouter(*slog.Logger) *chi.Mux`):
  - Middleware order is significant: `RequestID` → `SupressNotFound` → `CleanPath` → otelchi → otelchimetric (request duration, active requests, response body size) → `Recoverer`.
  - otelchi filter blocklist (no spans): `/health/liveness`, `/health/readiness`, `/docs`, `/openapi.yaml`, `/openapi.json`.
  - Metrics base config: name `"finsplitter"`, meter provider from `otel.GetMeterProvider()`.
  - The logger parameter is currently unused (`_`) — keep signature stable until callers change.
- Changes must keep the router usable by `v1/` (it attaches its own auth middleware + huma on top).

## Work Guidance

1. Router changes are rare; always re-read `NewRouter` before touching middleware order.
2. New global middleware → add to `r.Use(...)` in the correct position (earlier = runs first).
3. Request ID changes → update `request_id_test.go` (table-driven).

Example — read the request ID in a handler:

```go
logger := logctx.FromCtx(ctx)
logger.InfoContext(ctx, "handled", slog.String("request_id", logctx.GetRequestID(ctx)))
```

## 📂 Codebase References

- `internal/gateways/http/global_router.go`
- `internal/gateways/http/request_id.go`
- `internal/gateways/http/request_id_test.go`
- `pkg/logctx/request_id.go` — context key + With/Get helpers
- `internal/gateways/http/v1/api.go` — consumer of this router

## Verification

```bash
make test
make code-check
go test ./internal/gateways/http/...
```

## Child DOX Index

- `internal/gateways/http/v1/AGENTS.md` — Huma v2 API setup, handler groups, auth middleware (critical)
- This doc owns the router + request ID layer only; global standards live in the root AGENTS.md.
