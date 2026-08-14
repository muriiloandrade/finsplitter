<!-- Context: pkg/logctx | Priority: high | Version: 1.0 | Updated: 2026-08-12 -->

# pkg/logctx — context-aware logging + request ID

**Purpose**: Fills the gap where log/slog has no FromContext/NewContext helpers: stores/retrieves a `*slog.Logger` in the context and propagates a request ID. `logctx.FromCtx(ctx)` is the single sanctioned way to get a logger anywhere in the codebase. Mirrors the API of the replaced `github.com/veqryn/slog-context`.

## Ownership

- `pkg/logctx/logctx.go` — `NewCtx`, `FromCtx` (private `ctxKey struct{}`)
- `pkg/logctx/request_id.go` — `RequestIDKey`, `WithRequestID`, `GetRequestID`
- `pkg/logctx/logctx_test.go`, `pkg/logctx/request_id_test.go` — tests
- NOT: log format/level config (`pkg/telemetry/logging/logger.go`), request ID generation middleware (`internal/gateways/http/request_id.go`), OTel log bridge, or any direct logging — this package only stores and retrieves.

## Local Contracts

- `FromCtx(ctx)` returns the stored logger, or `slog.Default()` when absent or `ctx == nil`. `NewCtx(parent, logger)` falls back to `context.Background()` when `parent == nil`.
- The logger key is a private, unexported struct type (`ctxKey struct{}`) — context values can never collide; do not expose raw access to it.
- Request ID uses a typed key const `RequestIDKey ctxKeyRequestID = 0`; access only via `WithRequestID` / `GetRequestID` (empty string when absent or `ctx == nil`).
- This package never constructs or configures loggers — it stores whatever it is given.
- `pkg/telemetry/logging/context.go` (`CustomHandler`) reads the request ID via `logctx.GetRequestID(ctx)` to stamp `request_id` onto every log record.

## Work Guidance

Never call `slog.Info` / `slog.Error` directly — always resolve the logger from ctx:

```go
logger := logctx.FromCtx(ctx)
logger.InfoContext(ctx, "card brand created", slog.String("id", brand.ID))
```

Request ID propagation (middleware writes at the request boundary; consumers just read):

```go
ctx = logctx.WithRequestID(ctx, reqID)   // at request boundary
reqID := logctx.GetRequestID(ctx)         // anywhere downstream
```

Adding a new context value: copy the typed-key pattern above (private key type + exported accessor funcs). Keep behavior nil-safe — a nil ctx must never panic.

## 📂 Codebase References

- `pkg/logctx/logctx.go`, `pkg/logctx/request_id.go` — implementation
- `internal/gateways/http/request_id.go` — request ID generation + injection middleware
- `pkg/telemetry/logging/context.go` — `CustomHandler` adds `request_id` attr to records
- `pkg/logctx/logctx_test.go`, `pkg/logctx/request_id_test.go` — tests

## Verification

```bash
GOEXPERIMENT=jsonv2 go test ./pkg/logctx/...
make test
make code-check
```

## Child DOX Index

No child AGENTS.md files needed — leaf package. Root AGENTS.md owns global standards.
