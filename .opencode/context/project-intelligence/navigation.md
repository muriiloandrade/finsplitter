<!-- Context: project-intelligence/navigation | Priority: high | Version: 1.3 | Updated: 2026-06-27 -->

# Project Intelligence

| File | Description | Priority |
|------|-------------|----------|
| technical-domain.md | Tech stack, architecture, core patterns | critical |
| golang-patterns.md | Go-specific patterns, testing, security | critical |
| database-patterns.md | PostgreSQL, sqlc, migrations | high |
| docker-patterns.md | Containerization, multi-stage builds | medium |
| workflow.md | Development workflow, code quality | high |

## Quick Overview
- **Language**: Go 1.26 (GOEXPERIMENT=jsonv2 required)
- **Framework**: Huma v2 (REST API + OpenAPI)
- **Database**: PostgreSQL (pgx)
- **Auth**: Logto + jwx/jwkfetch (JWT, JWKS caching)
- **Cache**: Valkey 9.1 + go-redis/v9
- **Architecture**: Clean Architecture (Ports & Adapters)
- **Observability**: OpenTelemetry

## Update Frequency
Run `/add-context --update` when:
- Tech stack changes
- New patterns adopted
- Architecture decisions made
