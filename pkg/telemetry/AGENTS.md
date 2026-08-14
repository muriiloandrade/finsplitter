<!-- Context: pkg/telemetry | Priority: high | Version: 1.0 | Updated: 2026-08-12 -->

# pkg/telemetry — OpenTelemetry setup

**Purpose**: Central OTel v1.45.0 bootstrap: shared functional options, resource detection, shutdown contract, and one provider per concern in the sub-packages `tracing/`, `metrics/`, `logging/`. Project instrumentation stack: otelchi (server), otelhttp (client), otelpgx (DB), otelslog (log bridge), b3 + W3C propagators.

## Ownership

- `pkg/telemetry/options.go` — `Options` (private fields + getters), `Option`, `NewOptions`, `With*` options
- `pkg/telemetry/resource.go` — `NewResource` (process/OS/container/host auto-detection + service attrs)
- `pkg/telemetry/shutdown.go` — `ShutdownFunc func(context.Context) error`
- `pkg/telemetry/tracing/` — `tracer.go` (`NewTracerProvider`), `context.go` (`StartSpan`, `SpanFromContext`)
- `pkg/telemetry/metrics/` — `meter.go` (`NewMeterProvider`)
- `pkg/telemetry/logging/` — `logger.go` (`NewContextWithLogger`), `otel.go` (`NewLoggerProvider`), `context.go` (`CustomHandler`), `multihandler.go` (`MultiHandler`)
- NOT: app-specific spans/metrics/log attrs (each package owns its telemetry calls), config values (`internal/config`), middleware wiring (`internal/gateways/http`)

## Local Contracts

- All providers are built from `telemetry.Options` obtained via `NewOptions(opts...)` — `WithServiceName` and `WithExporterURL` are required (error otherwise). Defaults: version "local", env "development", exporter timeout 10s, metric export interval 60s.
- OTLP HTTP exporters only (gzip compression, 500ms initial retry, timeout from options; `WithInsecure(true)` for local dev only).
- Each `New*Provider` sets the GLOBAL provider (`otel.SetTracerProvider` / `otel.SetMeterProvider` / `global.SetLoggerProvider`) — they are process singletons; bootstrap once in `cmd/api`.
- Propagators set once in `tracing`: W3C TraceContext + Baggage + b3 (single + multiple header).
- Every provider returns a `telemetry.ShutdownFunc`; callers MUST invoke it on app exit to flush telemetry.
- Dependency direction: sub-packages import the parent (Options/Resource/ShutdownFunc); the parent NEVER imports sub-packages.
- `logging.NewContextWithLogger` takes `config.Config` — the only sub-package coupled to `internal/config`; bootstrap-only, called from `cmd/api`.
- Logging: `CustomHandler` injects `request_id` from `logctx`; `MultiHandler` fans records out to local output + OTel (`otelslog` bridge).
- Providers log via `logctx.FromCtx(ctx)`.

## Work Guidance

Bootstrap order in `cmd/api`: build `Options` → configure slog logger (returns ctx with logger) → create providers, collecting shutdown funcs → `defer` them in reverse order on exit.

```go
opts, err := telemetry.NewOptions(
    telemetry.WithServiceName("finsplitter"),
    telemetry.WithExporterURL("otel-collector:4318"),
    telemetry.WithInsecure(true), // local dev only
)
if err != nil { return err }
shutdown, err := tracing.NewTracerProvider(ctx, opts, 1.0) // 1.0 = always sample
if err != nil { return err }
defer func() { _ = shutdown(context.Background()) }()
```

To add a provider sub-package: follow the tracing/metrics/logging shape — `NewXProvider(ctx, opts) (…, telemetry.ShutdownFunc, error)`, set the global provider, log via `logctx.FromCtx(ctx)`, and document it here (Option B — the parent doc owns sub-packages). Sampler: always `ParentBased`; ratio 0.0 defaults to 1.0, values outside [0, 1] are rejected.

## 📂 Codebase References

- `pkg/telemetry/options.go`, `pkg/telemetry/resource.go`, `pkg/telemetry/shutdown.go`
- `pkg/telemetry/tracing/tracer.go`, `pkg/telemetry/tracing/context.go`
- `pkg/telemetry/metrics/meter.go`
- `pkg/telemetry/logging/logger.go`, `pkg/telemetry/logging/otel.go`, `pkg/telemetry/logging/context.go`, `pkg/telemetry/logging/multihandler.go`
- `pkg/httpclient/client.go` — otelhttp instrumentation (client side)
- `cmd/api/main.go` — bootstrap wiring

## Verification

```bash
GOEXPERIMENT=jsonv2 go test ./pkg/telemetry/...
make test
make code-check
```

## Child DOX Index

No separate AGENTS.md for sub-packages — per Option B the parent doc covers `tracing/`, `metrics/`, `logging/`. Root AGENTS.md owns global standards.
