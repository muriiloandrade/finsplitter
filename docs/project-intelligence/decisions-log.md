<!-- Context: project-intelligence/decisions | Priority: medium | Version: 1.1 | Updated: 2026-08-13 -->

# Decisions Log

**Purpose**: Major architectural and technical decisions with rationale, so agents/developers understand *why* the system is shaped this way. Newest first.

## Architecture & Stack

| Decision | Rationale | Impact |
|----------|-----------|--------|
| Clean Architecture (Ports & Adapters), modular monolith | Isolate business logic from infra (DB, HTTP) | Easy to test, swap DB/HTTP; use cases depend on interfaces |
| Huma v2 for REST API | OpenAPI spec generated from Go types/handler signatures | No separate spec maintenance |
| sqlc over ORM | Type-safe queries, zero runtime overhead | Every query is reviewed SQL in `sqlc/queries/*.sql` |
| pgx v5 over database/sql | Native PostgreSQL, built-in pooling | pgxpool with otelpgx tracing |
| OpenTelemetry for observability | Vendor-neutral traces/metrics/logs | Swap exporters without code changes |
| Logto over Auth0/Firebase | Self-hosted, OIDC-compliant | Full control, no vendor lock-in |
| Valkey over Redis | Drop-in Redis replacement, OSS license-safe | Same API (go-redis/v9), self-hosted |
| Multi-stage Docker build | Minimal production image (distroless) | Small attack surface, fast deploys |

## Auth-Specific

| Decision | Rationale | Impact |
|----------|-----------|--------|
| OAuth2 Device Authorization Grant for login (#207) | Passwordless flow fits mobile/SPA without redirects | `device_auth → poll → refresh` flow; Native App client |
| Token validation: JWT+JWKS with UserInfo fallback | Opaque tokens from device flow aren't JWTs | Two-strategy auth middleware |
| JWKS cached in Valkey (15min TTL) | Avoid hammering Logto; graceful degradation on cache failure | `cache` pkg + `jwkFetcher` interface |
| M2M client credentials for management API | Service-level access to Logto (create users) | Token cache with 60s safety buffer |
| Device token revocation (#219) | Security: users can revoke sessions | `/auth/device/revoke` endpoint + `device_revoke.go` use case |
| Password auth disabled | Security: device grant only | No password storage, no reset flow |
| `GOEXPERIMENT=jsonv2` required | jwx v4 depends on stdlib encoding/json/v2 | Must set env for all go build/test/run |

## Cross-Cutting

| Decision | Rationale | Impact |
|----------|-----------|--------|
| Request ID propagation, slog-context removal (#231) | Trace requests end-to-end; context-based logging | `pkg/logctx` + `internal/gateways/http/request_id.go` |
| ardanlabs/conf for env config | Declarative env parsing with struct tags | `conf` tags in `internal/config/config.go` |
| gofrs/uuid/v5 for entity IDs | UUIDs avoid enumeration, merge-friendly | pgx type adapter (`pgx-gofrs-uuid`) + sqlc override |
| shopspring/decimal for money math | High-precision decimal arithmetic | Entity amounts stored as decimal, not float |
| testcontainers-go for integration tests | Test against real Postgres/Valkey | `-short` skips; CI runs full suite |

## Proposed / Pending

| Proposal | Status | Detail |
|----------|--------|--------|
| `ardanlabs/conf` → `knadh/koanf/v2` | Proposed | See `docs/plans/conf-to-koanf-migration.md`; koanf already indirect dep |
| `gofrs/uuid/v5` → stdlib `uuid` (Go 1.27) | Proposed | See `docs/plans/uuid-stdlib-migration.md`; removes 2 direct + 1 indirect deps |

## 📂 Codebase References

- **Stack & architecture**: `technical-domain.md`, `business-tech-bridge.md`
- **DI wiring**: `cmd/api/main.go`
- **Config**: `internal/config/config.go`
- **Plans**: `docs/plans/` (conf→koanf, uuid→stdlib)
