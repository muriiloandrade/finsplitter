<!-- Context: pkg/httpclient | Priority: high | Version: 1.0 | Updated: 2026-08-12 -->

# pkg/httpclient — resty v3 wrapper

**Purpose**: Reusable, configurable HTTP client built on resty v3 (resty.dev/v3 v3.0.0-rc.3) with sane defaults, a functional-options API, and OpenTelemetry tracing/metrics via otelhttp on every client. Used by the Logto M2M client and device flow.

## Ownership

- `pkg/httpclient/client.go` — `Client`, `Option`, `New`, `Close`, `R`, `Resty`, `restyLogger`, all `With*` options, default constants
- NOT: upstream-specific endpoints, retry policies, or response/error types (the Logto client owns those)
- NOT: business error mapping — consumers parse responses (see Work Guidance)

## Local Contracts

- Constructed only via `New(opts...)` — zero value is not usable. Callers MUST call `Close()` to release resty internals.
- Defaults (override via options): retry count 3, retry wait 500ms → 10s exponential backoff with jitter, timeout 30s, retry on status 429/5xx/network errors (0).
- OTel is applied LAST in `New` via `otelhttp.NewTransport` over the (possibly option-set) transport — a custom `WithTransport` is still instrumented. `WithOpenTelemetry()` exists but is redundant with the default. Providers come from the global OTel SDK (`otel.GetTracerProvider` / `otel.GetMeterProvider`).
- `R(ctx)` returns a fresh `*resty.Request` wired with `ctx` (cancellation + deadline propagation).
- Resty's logger bridges to `*slog.Logger` via `restyLogger` — set it with `WithLogger(logctx.FromCtx(ctx))`.
- Response parsing contract for consumers: set `SetResult` (2xx) AND `SetResultError` (non-2xx), then check `resp.IsStatusFailure()`.

## Work Guidance

Follow the resty v3 error-parsing pattern (used by the Logto device flow):

```go
var result DeviceTokenResponse
resp, err := c.R(ctx).
    SetFormData(formData).
    SetResult(&result).
    SetResultError(&result). // Required: populates result.Error on 4xx
    Post(c.cfg.OIDCEndpoint + "/token")
if resp.IsStatusFailure() {
    switch result.Error {
    case "authorization_pending":
        return nil, ErrDeviceCodePending
    }
}
```

To add an option: declare `WithX(...)` returning `Option` that calls the matching `c.resty.SetX` method — mirror resty v3's method names. Never build a bare `resty.Client` in consumers; use `Resty()` only for advanced config not covered by options.

## 📂 Codebase References

- `pkg/httpclient/client.go` — client, options, logger adapter
- `internal/gateways/logto/m2m_client.go`, `internal/gateways/logto/device_flow.go` — consumers (device flow shows SetResult/SetResultError parsing)

## Verification

```bash
GOEXPERIMENT=jsonv2 go test ./pkg/httpclient/...
make test
make code-check
```

## Child DOX Index

No child AGENTS.md files needed — leaf package. Root AGENTS.md owns global standards.
