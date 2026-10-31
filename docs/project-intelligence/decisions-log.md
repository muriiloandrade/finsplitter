<!-- Context: project-intelligence/decisions | Priority: medium | Version: 1.2 | Updated: 2026-10-31 -->

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
| Go 1.27 toolchain (`GOEXPERIMENT` retired) | Go 1.27 stabilizes stdlib `encoding/json/v2`, the jwx v4 dependency | Plain `go build`/`go test`/`go run` — no env flag (previously `GOEXPERIMENT=jsonv2`) |

## Cross-Cutting

| Decision | Rationale | Impact |
|----------|-----------|--------|
| Request ID propagation, slog-context removal (#231) | Trace requests end-to-end; context-based logging | `pkg/logctx` + `internal/gateways/http/request_id.go` |
| ardanlabs/conf for env config (superseded) | Declarative env parsing with struct tags; migration to `knadh/koanf/v2` executed (2026-10-31) | `koanf` tags in `internal/config/config.go` (replaces `conf` tags) |
| stdlib `uuid` for entity IDs (was gofrs/uuid/v5) | UUIDs avoid enumeration + are merge-friendly; Go 1.27's first-party package drops 2 direct + 1 indirect deps | sqlc override (`import: uuid`) + native pgx v5 encode/scan — no `pgx-gofrs-uuid` adapter |
| shopspring/decimal for money math | High-precision decimal arithmetic | Entity amounts stored as decimal, not float |
| testcontainers-go for integration tests | Test against real Postgres/Valkey | `-short` skips; CI runs full suite |

## Completed

| Change | Date | Detail |
|--------|------|--------|
| `gofrs/uuid/v5` → stdlib `uuid` | 2026-10-11 | Unblocked by mockery v3.8.0 (v3.7.3 could not parse go1.27 modules); `docs/plans/uuid-stdlib-migration.md` executed in full, including the `GOEXPERIMENT=jsonv2` removal |
| `ardanlabs/conf/v3` → `knadh/koanf/v2` | 2026-10-31 | `docs/plans/conf-to-koanf-migration.md` executed in full; koanf v2.3.7 + `koanf/providers/env` direct, `ardanlabs/conf/v3` removed |

## 📂 Codebase References

- **Stack & architecture**: `technical-domain.md`, `business-tech-bridge.md`
- **DI wiring**: `cmd/api/main.go`
- **Config**: `internal/config/config.go`
- **Plans**: `docs/plans/` (conf→koanf)
