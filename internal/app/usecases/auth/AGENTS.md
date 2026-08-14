<!-- Context: app/usecases/auth | Priority: critical | Version: 1.1 | Updated: 2026-08-13 -->

# Auth Use Cases — identity lifecycle: register, device flow, me

**Purpose**: Owns the business logic for the identity lifecycle — passwordless registration, the Logto device authorization flow (request → poll → refresh → revoke), and current-user status (`/auth/me`). Orchestrates the Logto gateway and the user repository; no HTTP and no DB here.

## Ownership

- `interfaces.go` — consumer-defined gateway interfaces: `LogtoUserCreator`, `LogtoDeviceFlowClient` + compile-time checks (`var _ LogtoDeviceFlowClient = (*logto.Client)(nil)`)
- `register.go` — `RegisterUseCase`: creates Logto user (M2M) + local ID-only user record
- `device_auth.go` — `RequestDeviceAuthUseCase`: asks Logto for a device code + user code
- `device_poll.go` — `PollDeviceTokenUseCase`: polls for OIDC tokens after user approval
- `device_refresh.go` — `RefreshDeviceTokenUseCase`: exchanges refresh token for new tokens (rotation)
- `device_revoke.go` — `RevokeDeviceTokenUseCase`: revokes a refresh token (new in #219)
- `me.go` — `MeUseCase`: decides `NeedsSetup` from existence of a local record
- `mocks.gen.go` — mockery-generated mocks (`NewMockLogtoUserCreator`, `NewMockLogtoDeviceFlowClient`)
- Tests — `register_test.go`, `device_auth_test.go`, `device_poll_test.go`, `device_refresh_test.go`, `device_revoke_test.go`, `me_test.go` (all `package auth_test`)
- NOT: HTTP handlers (`internal/gateways/http/v1/auth/`), Logto client impl (`internal/gateways/logto/`), DB repos (`internal/gateways/postgres/`)

## Local Contracts

- **Consumer-defined interfaces**: this package declares what it needs from the Logto gateway; `internal/gateways/logto` satisfies it. Never import concrete `*logto.Client` into use case structs — depend on `LogtoUserCreator` / `LogtoDeviceFlowClient` only.
- **Handlers use concrete types**: `*auth.RegisterUseCase`, `*auth.MeUseCase`, etc. — no handler-layer interfaces (see handler pattern docs).
- **Local user storage** via `ports.UserRepository` (Create, GetByLogtoUserID, ExistsByLogtoUserID).
- **Logto is the source of truth** for identity. If Logto succeeds but the local record fails, do NOT roll back Logto — return the error and let the caller retry.
- **Error mapping** (register): `logto.ErrUserExists` → `errs.ErrUsernameTaken`; `logto.ErrEmailAlreadyInUse` / `errs.ErrDuplicate` → `errs.ErrUserAlreadyExists`. Wrap unknowns with `fmt.Errorf("...: %w", err)`.
- **Device flow errors pass through** from `internal/gateways/logto/errors.go`: `ErrDeviceCodePending` (retryable), `ErrDeviceCodeExpired` / `ErrDeviceCodeAccessDenied` (terminal).
- **Refresh token rotation**: Logto rotates refresh tokens — the caller MUST store the `refresh_token` returned by `RefreshDeviceTokenUseCase`.
- **Input validation**: empty email / device code / refresh token → `errs.ErrInvalidInput` (trim + check).
- **Never log tokens**; log via `logctx.FromCtx(ctx)`.

## Work Guidance

1. Device flow lifecycle (use case layer): `device_auth` (request code) → user approves in browser → `device_poll` (get tokens) → `device_refresh` (rotate) → `device_revoke` (revoke).
2. New auth use case: add consumer interface to `interfaces.go` → implement in the gateway → concrete UC struct + constructor → wire in `cmd/api/main.go` → regenerate mocks (`make generate`) → external-package table-driven test.
3. New method on `LogtoDeviceFlowClient`: implement it in `internal/gateways/logto/device_flow.go`; the compile-time checks in `interfaces.go` keep them honest.

Example — error mapping (register):

```go
logtoUser, err := uc.logtoM2M.CreateUser(ctx, username, "", input.Name, input.Email)
if errors.Is(err, logto.ErrUserExists) {
    return nil, errs.ErrUsernameTaken
}
```

## 📂 Codebase References

- `internal/app/usecases/auth/interfaces.go` — consumer-defined interfaces + compile-time checks
- `internal/app/usecases/auth/register.go`, `device_auth.go`, `device_poll.go`, `device_refresh.go`, `device_revoke.go`, `me.go`
- `internal/gateways/logto/device_flow.go`, `errors.go` — gateway implementations + sentinels
- `internal/app/ports/user_repo.go` — `UserRepository`
- `internal/domain/errs/errs.go` — domain error sentinels

## Verification

```bash
make test
make code-check
make generate   # regenerate mocks.gen.go after interface changes
go test ./internal/app/usecases/auth/...
```

## Child DOX Index

- No child AGENTS.md files needed — this is a leaf package. Gateway contracts live in `internal/gateways/logto/AGENTS.md`; global standards live in the root AGENTS.md.
