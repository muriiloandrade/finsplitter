<!-- Context: internal/config | Priority: high | Version: 1.2 | Updated: 2026-10-31 -->

# internal/config — Environment configuration

**Purpose**: Loads all runtime configuration from environment variables using `knadh/koanf/v2` struct tags. Single source of truth for env var names, defaults, and required values. Consumed only by `cmd/api`.

> Migration executed (2026-10-31): configuration now loads via `knadh/koanf/v2` (env provider + defaults + `validate`); `ardanlabs/conf/v3` was removed. Keep `koanf` conventions.

## Ownership

- `internal/config/config.go` — `Config`, section structs, `LoadEnv`
- `internal/config/config_test.go` — loader behavior tests
- Must NOT: read files, touch network/services, log secrets, or be imported outside `cmd/api`

## Local Contracts

- Flat env var names, no prefix: `APP_PORT`, `PG_USER`, `LOGTO_OIDC_ENDPOINT`, `OTEL_ENABLED`, ...
- Struct tags: `koanf:"key"` (e.g. `koanf:"db.pool.max_conns"`); required vars are validated in `Config.validate` (`PG_USER`, `PG_PASS`, `PG_HOST`, `PG_DB`)
- `LoadEnv(buildTag, buildCommit, buildTime)` returns `nil` when loading or validation fails → caller (`main`) panics
- Build info (`BuildTag`/`BuildCommit`/`BuildTime`) set after parse from ldflags — not env-driven
- Secrets (`PG_PASS`, `LOGTO_MGMT_CLIENT_SECRET`, `LOGTO_APP_CLIENT_SECRET`) must never be logged or printed
- Sensible defaults for everything non-required (DB pool, OTel, Redis URL)

## Work Guidance

Add a setting to the right section struct with a `koanf` tag; add a default if optional:

```go
type Database struct {
    // ...
    MaxReplicas int32 `koanf:"max_replicas"`
}
```

Required vars are enforced via `Config.validate`. Cover every new var in `config_test.go` (table-driven env → value assertions). The koanf migration landed; keep `koanf:"key"` tags and the env → koanf mapping in `envKeyToKoanfKey` in sync with the plan doc's Appendix A.

## Verification

```bash
go test -short ./internal/config/...
make test          # unit tests
make code-check    # format + lint
```

## Child DOX Index

No child AGENTS.md files needed — flat package.

## 📂 Codebase References

- `internal/config/config.go`
- `internal/config/config_test.go`
- `docs/plans/conf-to-koanf-migration.md` — executed koanf migration (2026-10-31)
