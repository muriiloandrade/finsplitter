<!-- Context: project-intelligence/technical | Priority: high | Version: 2.1 | Updated: 2026-08-13 -->

# Technical Domain

> Technical foundation, architecture, and key decisions. Source of truth for module-level rules: the **DOX AGENTS.md tree** (see `navigation.md`). This file is the quick-reference index.

## Quick Reference

- **Purpose**: Understand how the project works technically
- **Update When**: New features, refactoring, tech stack changes
- **Audience**: Developers, DevOps, technical stakeholders

## Primary Stack

| Layer | Technology | Version | Rationale |
|-------|-----------|---------|-----------|
| Language | Go | 1.26.4 | Static typing, fast builds, single binary |
| API Framework | Huma v2 | 2.39.1 | OpenAPI spec generated from Go types |
| HTTP Router | Chi (via Huma) | — | Idiomatic Go routing |
| Database | PostgreSQL | 18 | Relational model, JSONB, constraints |
| DB Access | pgx v5 + sqlc | — | Type-safe SQL, no ORM overhead |
| Cache | Valkey + go-redis/v9 | — | Redis-compatible, self-hosted |
| Identity/Auth | Logto (OIDC) + jwx/v4 + jwkfetch | — | Self-hosted, JWT+JWKS + device grant |
| Observability | OpenTelemetry + slog | — | Vendor-neutral traces/metrics/logs |
| HTTP Client | resty v3 (via `pkg/httpclient`) | — | Retry, timeout, OTel hooks |
| Config | ardanlabs/conf | v3 | Declarative env parsing (→ koanf proposed) |
| IDs | gofrs/uuid/v5 | — | UUIDs for entity IDs (→ stdlib proposed) |
| Money | shopspring/decimal | — | Precise monetary math |
| Testing | testify + mockery + testcontainers-go | — | Mocks + real-Postgres integration tests |
| Infrastructure | Docker + GitHub Actions | — | Multi-stage distroless builds, CI/CD |

## Architecture Pattern

```
Type: Modular Monolith
Pattern: Clean Architecture (Ports & Adapters / Hexagonal)
Diagram: see business-tech-bridge.md "Key Architectural Contracts"
```

```
HTTP Gateway (adapters) → Use Cases (core logic) → Ports (interfaces) → Postgres/Logto (adapters)
```

### Why This Architecture?

Business logic (use cases + domain) is isolated from infrastructure (HTTP, DB, Logto). Use cases depend on small consumer-defined interfaces, so they're unit-testable with mocks and the DB/HTTP layers can be swapped without touching business rules. Modular monolith keeps deployment simple (one binary) while preserving domain boundaries for a future split.

## Project Structure

```
finsplitter/
├── cmd/api/                  # Entry point, DI wiring
├── internal/
│   ├── app/
│   │   ├── ports/            # Repository interfaces (contracts)
│   │   └── usecases/         # Business logic: auth, card-brand, profile
│   ├── config/               # Env config (ardanlabs/conf)
│   ├── domain/               # Entities, domain errors, transactioner
│   └── gateways/
│       ├── http/             # Router, request ID middleware
│       ├── http/v1/          # Huma handlers, routes, middleware
│       ├── logto/            # Logto M2M + device flow clients
│       └── postgres/         # pgx pool, sqlc, migrations, testutils
├── pkg/
│   ├── cache/                # Valkey/Redis client
│   ├── httpclient/           # Resty v3 wrapper
│   ├── logctx/               # Context-aware logging + request ID
│   └── telemetry/            # OTel tracing, metrics, logging
├── docs/                     # plans/, project-intelligence/ (domain spec)
└── .github/                  # CI/CD
```

**Key Directories**: each directory owns an `AGENTS.md` with its local contracts (see `navigation.md` DOX Tree). Business rules per module live in `internal/app/usecases/*/AGENTS.md`.

## Key Technical Decisions

| Decision | Rationale | Impact |
|----------|-----------|--------|
| Clean Architecture, modular monolith | Isolate business logic from infra | Testable, swappable adapters |
| Huma v2 (OpenAPI from code) | No separate spec maintenance | Spec always matches handlers |
| sqlc over ORM | Type-safe, zero-runtime queries | Reviewed SQL in `sqlc/queries/` |
| Logto over Auth0/Firebase | Self-hosted, OIDC | No vendor lock-in |
| Device Authorization Grant (passwordless) | Mobile/SPA login without redirects | Auth: register, device flow, me, refresh, revoke |
| JWT+JWKS + UserInfo fallback | Opaque device tokens aren't JWTs | Two-strategy auth middleware |
| Valkey over Redis | OSS license-safe drop-in | Same API, self-hosted |
| OpenTelemetry | Vendor-neutral | Swap exporters freely |
| GOEXPERIMENT=jsonv2 (jwx v4 dep) | Required by dependency | Set for all go build/test/run |

Full history with alternatives: `decisions-log.md`.

## Integration Points

| System | Purpose | Protocol | Direction |
|--------|---------|----------|-----------|
| Logto OIDC | Authentication (JWT/JWKS, device grant, UserInfo) | OIDC/HTTP | Inbound + Outbound |
| Logto Management API | Create/update users, M2M client credentials | REST (HTTP) | Outbound |
| PostgreSQL | Persistence (users, card_brands; future: person/card/bill/transaction) | PostgreSQL | Internal |
| Valkey | Cache: JWKS (15min TTL), M2M token (60s safety buffer) | RESP | Internal |
| OpenTelemetry Collector | Traces/metrics/logs export | OTLP | Outbound |

## Technical Constraints

| Constraint | Origin | Impact |
|------------|--------|--------|
| `GOEXPERIMENT=jsonv2` required | jwx v4 dep | Must prefix all go build/test/run |
| Passwordless only (device grant) | Security decision | No password auth/reset flow |
| DB SSL required in prod | Security (`PG_SSL_MODE=require`) | TLS for all DB connections |
| `LOGTO_APP_CLIENT_ID` must match API Resource aud | Logto config | Auth fails if mismatched |
| SQL-injection safety | sqlc parameterized queries | No raw SQL concat |
| No sensitive data in logs | Security | Tokens/secrets never logged |
| UUID entity IDs + `last_modified_date` trigger | Data integrity | Consistent ID/history pattern |

## Development Environment

```
Setup: cp .env.example .env → configure → make start-infra → make start-dev
Requirements: Go 1.26.4+, Docker, GOEXPERIMENT=jsonv2
Local Dev: make start-dev (infra + hot reload)
Testing:  make test | go test ./... -cover
Generate: make generate (sqlc + mocks)
Quality:  make code-check (gofmt + golangci-lint)
```

## Deployment

```
Environment: Production (distroless image), compose for local/infra
Platform: Docker multi-stage (setup → builder → production)
Image: gcr.io/distroless/static-debian12:nonroot (non-root, HEALTHCHECK)
CI/CD: GitHub Actions (.github/)
Monitoring: OpenTelemetry → Collector → Grafana/Tempo/Loki (compose.monitoring.yml)
Compose: compose.yml (backend), compose.infra.yml (Postgres+Valkey+Logto), compose.monitoring.yml (OTel/Grafana/Tempo/Loki), compose.debug.yml
Security: make docker-scout (image scans), .dockerignore, no secrets in images
```

## Onboarding Checklist

- [x] Know the primary tech stack (Go, Huma v2, PostgreSQL, Logto, Valkey, OTel)
- [x] Understand the architecture pattern (Clean Architecture, modular monolith) and why
- [x] Know the key project directories and their purpose (DOX tree per directory)
- [x] Understand major technical decisions and rationale (`decisions-log.md`)
- [x] Know integration points (Logto, PostgreSQL, Valkey, OTel)
- [x] Set up local development environment (`.env.example` → `make start-dev`)
- [x] Know how to run tests (`make test`) and quality checks (`make code-check`)

## Related Files

- `navigation.md` — quick overview + DOX tree routing
- `business-domain.md` — why this technical foundation exists
- `business-tech-bridge.md` — how business needs map to technical solutions
- `decisions-log.md` — full decision history with context
- `living-notes.md` — active issues, debt, open questions
- `docs/plans/` — proposed migrations (conf→koanf, uuid→stdlib)
