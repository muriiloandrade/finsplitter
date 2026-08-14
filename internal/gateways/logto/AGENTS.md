<!-- Context: gateways/logto | Priority: critical | Version: 1.0 | Updated: 2026-08-12 -->

# Logto Gateway — M2M Management API client + device authorization flow

**Purpose**: Talks to Logto for identity: user CRUD via the Management API (M2M client-credentials grant with token caching) and the device authorization flow (request → poll → refresh → revoke) via the Native App client. Wraps the shared resty v3 `pkg/httpclient`; contains no business logic.

## Ownership

- `m2m_client.go` — `Client`, `Config`, `NewClient` (+ `ClientOption`s); M2M token caching (`getToken`, double-checked locking, 60s safety buffer); Management API ops `CreateUser` / `UpdateUser`
- `device_flow.go` — device flow ops `RequestDeviceCode`, `PollDeviceToken`, `RefreshDeviceToken`, `RevokeDeviceToken` + response/error structs
- `errors.go` — sentinel errors: `ErrM2MUnauthorized`, `ErrUserExists`, `ErrUserNotFound`, `ErrEmailAlreadyInUse`, `ErrAppClientNotConfigured`; device errors `ErrDeviceCodePending`, `ErrDeviceCodeExpired`, `ErrDeviceCodeAccessDenied`, `ErrDeviceTokenRevoked`
- Tests — `m2m_client_test.go`, `device_flow_test.go`
- NOT: consumer interfaces (`internal/app/usecases/auth/interfaces.go`), token validation middleware (`internal/gateways/http/v1/auth/`), HTTP client (`pkg/httpclient`)

## Local Contracts

- **One `Client` satisfies consumer interfaces**: `var _ auth.LogtoUserCreator = (*logto.Client)(nil)` and `var _ auth.LogtoDeviceFlowClient = (*logto.Client)(nil)` live in `internal/app/usecases/auth/interfaces.go` — new exported methods must keep them compiling.
- **M2M auth** (Management API): `grant_type=client_credentials` POST to `OIDCEndpoint + "/token"` with `resource` (default `ManagementBaseURL + "/api"`), `scope=all`, machine `ClientID`/`ClientSecret`. Token cached in memory under `sync.RWMutex`; refreshed `tokenExpiryBuffer` = 60s before expiry (clamped to 0 for sub-60s tokens). Token-endpoint status ≥ 400 → `ErrM2MUnauthorized`.
- **Audience contract**: `LOGTO_APP_CLIENT_ID` must match the Logto API Resource identifier (the `aud` claim) — validated by the auth middleware; keep env values in sync.
- **Device flow uses the Native App client**: all device ops send `client_id` = `DeviceAppClientID` (`LOGTO_DEVICE_APP_CLIENT_ID`) with `scope: openid profile email offline_access`; empty `DeviceAppClientID` → `ErrAppClientNotConfigured`. M2M and device credentials are NOT interchangeable.
- **resty v3 error parsing**: set `SetResult` (2xx) AND `SetResultError` (non-2xx) to the same struct — both are required or `result.Error`/`result.ErrorDescription` stay empty on 4xx (e.g. `PollDeviceToken` 400 → `authorization_pending` / `slow_down` / `expired_token` / `access_denied`).
- **Error mapping**: map HTTP status to sentinel errors; call sites compare with `errors.Is`. Unknown statuses → `fmt.Errorf("op: status %d", resp.StatusCode())` — no raw bodies in errors.
- **Token rotation**: Logto rotates refresh tokens — callers MUST persist the `refresh_token` returned by `RefreshDeviceToken`.
- **Graceful degradation**: token cache is best-effort in memory; on fetch failure return the wrapped error — never serve a stale token. Transient HTTP failures are absorbed by `pkg/httpclient` retries (3×).
- **No secrets in logs**: never log tokens, client secrets, or form data. OTel tracing, retries, timeout (10s) come from `pkg/httpclient` defaults; override via `WithHTTPClientOptions` (tests use custom transport).

## Work Guidance

1. New Management API op: add method in `m2m_client.go` → `getToken(ctx)` → resty `SetAuthToken(token)` → map status to sentinel error → add to a consumer interface + compile-time check in `auth/interfaces.go` → `make generate-mocks`.
2. New device flow op: add method in `device_flow.go`, form data with `DeviceAppClientID`, always both `SetResult` + `SetResultError`.
3. New failure mode → new sentinel in `errors.go`, wrap with `%w`, map it in the consuming use case.

Example — poll with both result bindings (4xx body required):

```go
var result DeviceTokenResponse
resp, err := c.httpClient.R(ctx).
    SetFormData(formData).
    SetResult(&result).
    SetResultError(&result). // populates result.Error on 4xx
    Post(c.cfg.OIDCEndpoint + "/token")
if resp.StatusCode() == http.StatusBadRequest && result.Error == "authorization_pending" {
    return nil, ErrDeviceCodePending
}
```

## 📂 Codebase References

- `internal/gateways/logto/m2m_client.go`, `device_flow.go`, `errors.go`
- `internal/app/usecases/auth/interfaces.go` — consumer-defined interfaces + compile-time checks
- `pkg/httpclient/` — resty v3 wrapper (retry, timeout, OTel)
- `internal/config/config.go` — Logto env config (`LOGTO_*`)

## Verification

```bash
make test
make code-check
go test ./internal/gateways/logto/...
make generate-mocks   # after interface changes in auth/interfaces.go
```

## Child DOX Index

- No child AGENTS.md files needed — leaf package. Consumer contract lives in `internal/app/usecases/auth/AGENTS.md`; global standards live in the root AGENTS.md.
