<!-- Context: internal/config | Priority: high | Version: 1.0 | Updated: 2026-08-12 -->

# internal/config — Environment configuration

**Purpose**: Loads all runtime configuration from environment variables using `ardanlabs/conf/v3` struct tags. Single source of truth for env var names, defaults, and required values. Consumed only by `cmd/api`.

> ⚠️ **Planned, not done**: a migration to `knadh/koanf/v2` is proposed in `docs/plans/conf-to-koanf-migration.md` (Status: Proposed). Keep `ardanlabs/conf` conventions until that plan is executed.

## Ownership

- `internal/config/config.go` — `Config`, section structs, `LoadEnv`
- `internal/config/config_test.go` — loader behavior tests
- Must NOT: read files, touch network/services, log secrets, or be imported outside `cmd/api`

## Local Contracts

- Flat env var names, no prefix: `APP_PORT`, `PG_USER`, `LOGTO_OIDC_ENDPOINT`, `OTEL_ENABLED`, ...
- Struct tags: `conf:"env:NAME,default:VALUE"`; required vars use `required` (`PG_USER`, `PG_PASS`, `PG_HOST`, `PG_DB`)
- `LoadEnv(buildTag, buildCommit, buildTime)` returns `nil` on any parse failure → caller (`main`) panics
- Build info (`BuildTag`/`BuildCommit`/`BuildTime`) set after parse from ldflags — not env-driven
- Secrets (`PG_PASS`, `LOGTO_MGMT_CLIENT_SECRET`, `LOGTO_APP_CLIENT_SECRET`) must never be logged or printed
- Sensible defaults for everything non-required (DB pool, OTel, Redis URL)

## Work Guidance

Add a setting to the right section struct with a `conf` tag; add a default if optional:

```go
type Database struct {
    // ...
    MaxReplicas int32 `conf:"env:PG_MAX_REPLICAS,default:0"`
}
```

Required vars get the `required` keyword. Cover every new var in `config_test.go` (table-driven env → value assertions). If the koanf migration lands, switch tags to `koanf:"key"` and update the mapping table in the plan doc.

## Verification

```bash
GOEXPERIMENT=jsonv2 go test -short ./internal/config/...
make test          # unit tests
make code-check    # format + lint
```

## Child DOX Index

No child AGENTS.md files needed — flat package.

## 📂 Codebase References

- `internal/config/config.go`
- `internal/config/config_test.go`
- `docs/plans/conf-to-koanf-migration.md` — planned koanf migration
