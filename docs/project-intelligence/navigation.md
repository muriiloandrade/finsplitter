<!-- Context: project-intelligence/navigation | Priority: high | Version: 2.3 | Updated: 2026-08-13 -->

# Project Intelligence — Navigation

> Start here for quick project understanding. Detailed technical context lives in the **DOX AGENTS.md tree** (per-module docs); this folder holds the knowledge index + cross-cutting references.

## Structure

```
docs/project-intelligence/
├── navigation.md              # This file — quick overview + routing
├── business-domain.md         # Business context, entities, rules, workflows
├── business-tech-bridge.md    # Business needs → technical implementation
├── technical-domain.md        # Index — stack + routing to DOX tree
├── decisions-log.md           # Major decisions with rationale
└── living-notes.md            # Active issues, debt, open questions
```

## Quick Routes

| What You Need | File |
|---------------|------|
| Understand the "why" | `business-domain.md` |
| Understand the "how" | `technical-domain.md` → DOX tree |
| See the connection | `business-tech-bridge.md` |
| Know the decision context | `decisions-log.md` |
| Current state / open issues | `living-notes.md` |
| **Module-local rules** | **Per-module `AGENTS.md` in the DOX tree** |

## Quick Overview

- **Language**: Go 1.26.4 (`GOEXPERIMENT=jsonv2` required)
- **Framework**: Huma v2.39.1 (REST + OpenAPI)
- **Database**: PostgreSQL 18 (pgx v5, sqlc)
- **Auth**: Logto + jwx/jwkfetch (JWT, JWKS caching, device authorization grant, UserInfo fallback)
- **Cache**: Valkey + go-redis/v9
- **Architecture**: Clean Architecture (Ports & Adapters), modular monolith
- **Observability**: OpenTelemetry v1.45

## DOX Tree (source of truth for module rules)

| Area | Doc | Priority |
|------|-----|----------|
| Entry, DI wiring | `cmd/api/AGENTS.md` | critical |
| Env config | `internal/config/AGENTS.md` | high |
| Entities, domain errors | `internal/domain/AGENTS.md` | critical |
| Repository interfaces | `internal/app/ports/AGENTS.md` | critical |
| Auth use cases | `internal/app/usecases/auth/AGENTS.md` | critical |
| Card brand use cases | `internal/app/usecases/card-brand/AGENTS.md` | high |
| Profile setup use case | `internal/app/usecases/profile/AGENTS.md` | high |
| Global router, request ID | `internal/gateways/http/AGENTS.md` | critical |
| Huma v2 handlers, middleware | `internal/gateways/http/v1/AGENTS.md` | critical |
| Logto M2M + device flow | `internal/gateways/logto/AGENTS.md` | critical |
| Postgres, sqlc, migrations | `internal/gateways/postgres/AGENTS.md` | critical |
| Valkey/Redis cache | `pkg/cache/AGENTS.md` | high |
| Resty v3 HTTP client | `pkg/httpclient/AGENTS.md` | high |
| Context logging + request ID | `pkg/logctx/AGENTS.md` | high |
| OTel setup | `pkg/telemetry/AGENTS.md` | high |

Global rules (naming, standards, security, testing) live in the root `AGENTS.md` (DOX rail).

## Integration

- Root `AGENTS.md` (DOX rail) — global rules + Child DOX Index
- `business-domain.md` — canonical domain spec (entities, business rules, workflows)
- `docs/plans/` — proposed migrations (conf→koanf, uuid→stdlib)

## Maintenance

- AGENTS.md files update via the DOX pass after every meaningful change (root AGENTS.md contract)
- Update `technical-domain.md` on stack changes or new modules
- Review `living-notes.md` regularly; archive resolved items into `decisions-log.md`
