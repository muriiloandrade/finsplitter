# Migration: `ardanlabs/conf/v3` → `knadh/koanf/v2`

> **Status**: Proposed
> **Dependencies**: `github.com/knadh/koanf/v2` (already an indirect dep, v2.3.5 — will become direct)
> **Dependencies removed**: `github.com/ardanlabs/conf/v3`

## 1. Why

`ardanlabs/conf/v3` is used only in `internal/config/config.go` for env-var parsing with struct tags. `knadh/koanf` is already in the module graph (indirect, pulled by huma tooling), is more actively maintained, and supports multiple providers (env, file, flags) with a single `Unmarshal` step. Migrating removes a direct dependency and unifies config loading under one library.

## 2. Current state

`internal/config/config.go` uses:

```go
type Config struct {
    App   Application
    Env   Environment
    DB    Database
    Redis Redis
    OTel  OpenTelemetry
    Logto Logto
}

type Application struct {
    Port        int    `conf:"env:APP_PORT,default:3033"`
    Name        string `conf:"env:APP_NAME,default:finsplitter"`
    Version     string `conf:"env:APP_VERSION,default:dev"`
    // ...
}

func LoadEnv(buildTag, buildCommit, buildTime string) *Config {
    var cfg Config
    if _, err := conf.Parse("", &cfg); err != nil {
        if errors.Is(err, conf.ErrHelpWanted) {
            return nil
        }
        slog.Default().Error("fail to load configurations", slog.Any("error", err))
        return nil
    }
    cfg.App.BuildCommit = buildCommit
    cfg.App.BuildTag = buildTag
    cfg.App.BuildTime = buildTime
    return &cfg
}
```

**Current env vars** (no prefix, flat names): `APP_PORT`, `APP_NAME`, `APP_VERSION`, `ENV_NAME`, `LOG_FORMAT`, `PG_USER`, `PG_PASS`, `PG_HOST`, `PG_PORT`, `PG_DB`, `PG_SSL_MODE`, `PG_SCHEMA`, `PG_MAX_CONNS`, `PG_MIN_CONNS`, `PG_MAX_CONN_LIFETIME`, `PG_MAX_CONN_IDLE_TIME`, `PG_HEALTH_CHECK_PERIOD`, `PG_CONNECT_TIMEOUT`, `REDIS_URL`, `LOGTO_*`, `OTEL_*`.

## 3. Target design (koanf)

### 3.1 Struct tags

Replace `conf:"env:X,default:Y"` with `koanf:"key"` tags. koanf maps keys to struct fields via `mapstructure`; nested structs map to dotted keys.

```go
type Config struct {
    App   Application `koanf:"app"`
    Env   Environment `koanf:"env"`
    DB    Database    `koanf:"db"`
    Redis Redis       `koanf:"redis"`
    OTel  OpenTelemetry `koanf:"otel"`
    Logto Logto       `koanf:"logto"`
}

type Application struct {
    Port    int    `koanf:"port"`
    Name    string `koanf:"name"`
    Version string `koanf:"version"`
    // Build* fields are NOT env-driven; set after unmarshal (see 3.3)
}
```

### 3.2 Loading

```go
package config

import (
    "log/slog"
    "strings"

    "github.com/knadh/koanf/parsers/rawbytes" // not needed for env-only
    "github.com/knadh/koanf/providers/env"
    "github.com/knadh/koanf/v2"
)

func LoadEnv(buildTag, buildCommit, buildTime string) *Config {
    k := koanf.New(".")

    // Load defaults first (so env overrides them).
    if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
        slog.Default().Error("fail to load default config", slog.Any("error", err))
        return nil
    }

    // Load env vars. No prefix — transform APP_PORT -> app.port, PG_MAX_CONNS -> db.max_conns.
    if err := k.Load(env.Provider(".", env.Opt{
        TransformFunc: func(key, value string) (string, any) {
            key = strings.ToLower(key)
            key = strings.ReplaceAll(key, "_", ".")
            return key, value
        },
    }), nil); err != nil {
        slog.Default().Error("fail to load env config", slog.Any("error", err))
        return nil
    }

    var cfg Config
    if err := k.Unmarshal("", &cfg); err != nil {
        slog.Default().Error("fail to unmarshal config", slog.Any("error", err))
        return nil
    }

    cfg.App.BuildCommit = buildCommit
    cfg.App.BuildTag = buildTag
    cfg.App.BuildTime = buildTime
    return &cfg
}
```

### 3.3 Defaults map

```go
var defaults = map[string]any{
    "app.port":    3033,
    "app.name":    "finsplitter",
    "app.version": "dev",
    "env.name":    "local",
    "env.log_format": "text",
    "db.port":     5432,
    "db.ssl_mode": "require",
    "db.schema":   "public",
    "db.pool.max_conns": 10,
    "db.pool.min_conns": 1,
    "db.pool.max_conn_lifetime": "1h",
    "db.pool.max_conn_idle_time": "10m",
    "db.pool.health_check_period": "1m",
    "db.pool.connect_timeout": "15s",
    "redis.url": "redis://localhost:6379/0",
    "logto.oidc_endpoint": "http://localhost:3001/oidc",
    "logto.issuer": "http://localhost:3001/oidc",
    "logto.endpoint": "http://localhost:3001",
    "otel.enabled": false,
    "otel.service_name": "finsplitter",
    "otel.exporter_otlp_endpoint": "http://localhost:4318",
    "otel.exporter_insecure": true,
    "otel.enable_traces": true,
    "otel.enable_metrics": true,
    "otel.enable_logs": true,
    "otel.sampler_ratio": 1.0,
    "otel.exporter_timeout": "30s",
    "otel.export_interval": "5s",
}
```

### 3.4 Required fields

`ardanlabs/conf` supports `required`; koanf does **not** have a built-in required marker. Add explicit validation after unmarshal:

```go
func (c *Config) validate() error {
    var missing []string
    if c.DB.User == "" { missing = append(missing, "PG_USER") }
    if c.DB.Password == "" { missing = append(missing, "PG_PASS") }
    if c.DB.Host == "" { missing = append(missing, "PG_HOST") }
    if c.DB.DBName == "" { missing = append(missing, "PG_DB") }
    if len(missing) > 0 {
        return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
    }
    return nil
}
```

### 3.5 Help/usage output

`ardanlabs/conf` provided `conf.ErrHelpWanted` + `conf.Usage`. koanf has no equivalent. If the `-h`/usage behavior is needed, implement it manually (e.g., a `--help` flag via `flag` or a printed defaults table). **Recommendation**: drop it — the app is env-configured, and `main.go` currently treats `nil` config as fatal anyway.

## 4. Changes needed

| File | Change |
|------|--------|
| `internal/config/config.go` | Rewrite `LoadEnv` with koanf; add `defaults` map; add `validate()`; replace all `conf:` tags with `koanf:` tags |
| `internal/config/config_test.go` | Update tests to match new loader behavior (defaults + env override semantics should be identical — verify) |
| `go.mod` | `knadh/koanf/v2` moves from indirect → direct; remove `ardanlabs/conf/v3`; `go mod tidy` |
| `cmd/api/main.go` | No change expected (still calls `config.LoadEnv`) |

## 5. Possible drawbacks / risks

| # | Risk | Severity | Mitigation |
|---|------|----------|------------|
| 1 | **Env var name mapping** — `PG_MAX_CONNS` → `db.max_conns` requires a `_`→`.` transform; ambiguous names (e.g., `LOG_FORMAT` → `env.log_format` vs `log.format`) must be verified | Medium | Enumerate every env var in a test that asserts the final `Config` values (extend `config_test.go`) |
| 2 | **No `required` support** — silently missing required vars would produce empty strings instead of a load error | High | Add explicit `validate()` (3.4) and a test for missing required vars |
| 3 | **No help/usage output** — `conf.ErrHelpWanted` behavior disappears | Low | Acceptable; document in README if needed |
| 4 | **Duration parsing** — koanf/mapstructure parses `time.Duration` from strings ("1h") — verify all `time.Duration` fields still parse | Medium | Covered by config tests |
| 5 | **Bool parsing** — `OTEL_ENABLED=false` must parse correctly (mapstructure handles "true"/"false") | Low | Covered by `TestOpenTelemetryConfig` |
| 6 | **koanf version** — v2.3.5 is already in the graph; confirm API stability for `env.Opt.TransformFunc` | Low | Pin to v2.3.5+ |

## 6. Execution order

1. Add `defaults` map + rewrite `LoadEnv` in `config.go` (keep `Config` struct shape identical so callers don't change).
2. Add `validate()` for required fields.
3. Update `config_test.go` — add a full env-var mapping table test.
4. `go mod tidy` (promote koanf to direct, drop ardanlabs/conf).
5. `make test` + `make code-check`.

## Appendix A — full env var → koanf key mapping

| Env var | koanf key |
|---|---|
| `APP_PORT` | `app.port` |
| `APP_NAME` | `app.name` |
| `APP_VERSION` | `app.version` |
| `ENV_NAME` | `env.name` |
| `LOG_FORMAT` | `env.log_format` |
| `PG_USER` | `db.user` |
| `PG_PASS` | `db.password` |
| `PG_HOST` | `db.host` |
| `PG_PORT` | `db.port` |
| `PG_DB` | `db.db_name` |
| `PG_SSL_MODE` | `db.ssl_mode` |
| `PG_SCHEMA` | `db.schema` |
| `PG_MAX_CONNS` | `db.pool.max_conns` |
| `PG_MIN_CONNS` | `db.pool.min_conns` |
| `PG_MAX_CONN_LIFETIME` | `db.pool.max_conn_lifetime` |
| `PG_MAX_CONN_IDLE_TIME` | `db.pool.max_conn_idle_time` |
| `PG_HEALTH_CHECK_PERIOD` | `db.pool.health_check_period` |
| `PG_CONNECT_TIMEOUT` | `db.pool.connect_timeout` |
| `REDIS_URL` | `redis.url` |
| `LOGTO_OIDC_ENDPOINT` | `logto.oidc_endpoint` |
| `LOGTO_ISSUER` | `logto.issuer` |
| `LOGTO_ENDPOINT` | `logto.endpoint` |
| `LOGTO_MGMT_CLIENT_ID` | `logto.mgmt_client_id` |
| `LOGTO_MGMT_CLIENT_SECRET` | `logto.mgmt_client_secret` |
| `LOGTO_MGMT_API_RESOURCE` | `logto.mgmt_api_resource` |
| `LOGTO_APP_CLIENT_ID` | `logto.app_client_id` |
| `LOGTO_APP_CLIENT_SECRET` | `logto.app_client_secret` |
| `LOGTO_DEVICE_APP_CLIENT_ID` | `logto.device_app_client_id` |
| `OTEL_ENABLED` | `otel.enabled` |
| `OTEL_SERVICE_NAME` | `otel.service_name` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `otel.exporter_otlp_endpoint` |
| `OTEL_EXPORTER_INSECURE` | `otel.exporter_insecure` |
| `OTEL_ENABLE_TRACES` | `otel.enable_traces` |
| `OTEL_ENABLE_METRICS` | `otel.enable_metrics` |
| `OTEL_ENABLE_LOGS` | `otel.enable_logs` |
| `OTEL_SAMPLER_RATIO` | `otel.sampler_ratio` |
| `OTEL_EXPORTER_TIMEOUT` | `otel.exporter_timeout` |
| `OTEL_EXPORT_INTERVAL` | `otel.export_interval` |

> ⚠️ The naive `_` → `.` transform produces `db.db_name` for `PG_DB` and `env.log_format` for `LOG_FORMAT`. These are correct **only if** the struct tags match exactly. The mapping table above must be encoded in tests to lock it in.