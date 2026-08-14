<!-- Context: app/usecases/profile | Priority: high | Version: 1.1 | Updated: 2026-08-13 -->

# Profile Use Cases — profile setup

**Purpose**: Single-use-case package — completes profile setup for a user authenticated via Logto: updates the Logto profile (username, name, phone, picture) and creates the local Finsplitter user record. No HTTP, no DB here.

## Ownership

- `setup.go` — `SetupUseCase`, `SetupInput` / `SetupOutput`
- `interfaces.go` — consumer-defined `LogtoUserUpdater` interface + compile-time check (`var _ LogtoUserUpdater = (*logto.Client)(nil)`)
- `mocks.gen.go` — mockery-generated `MockLogtoUserUpdater` (from `.mockery.yml`)
- `setup_test.go` — `package profile_test` (external)
- NOT: HTTP handlers (`internal/gateways/http/v1/profile/`), Logto client impl (`internal/gateways/logto/`), user repo impl (`internal/gateways/postgres/`)

## Local Contracts

- **Consumer-defined interface**: `LogtoUserUpdater` (UpdateUser) is declared in `interfaces.go` and satisfied by `*logto.Client` in production — the use case never imports the concrete client.
- Deps: `ports.UserRepository` + `LogtoUserUpdater`; constructor `NewSetupUseCase(userRepo, logtoClient)`.
- **Idempotency**: `ExistsByLogtoUserID` → true → return `errs.ErrDuplicate` (handler maps to 409). Setup is a one-time action.
- **Order matters**: update the Logto profile FIRST (best-effort), then create the local record. Logto is the source of truth — if the local create fails after a successful Logto update, return the error; the user can retry setup (idempotent).
- Input: empty `LogtoUserID` → `errs.ErrInvalidInput`.
- Never log token/profile internals; log via `logctx.FromCtx(ctx)`.

## Work Guidance

1. Flow: validate input → existence check (duplicate guard) → `UpdateUser` on Logto → `Create` local record → return `SetupOutput{UserID}`.
2. Tests: external package `profile_test`; `LogtoUserUpdater` mocked via mockery `profile.NewMockLogtoUserUpdater(t)`; `UserRepository` uses `ports.NewMockUserRepository(t)`.
3. Adding a profile field: extend `SetupInput`, the `UpdateUser` signature, the Logto gateway call, and mock expectations together.

Example — setup order:

```go
exists, err := uc.userRepo.ExistsByLogtoUserID(ctx, input.LogtoUserID)
if exists {
    return nil, errs.ErrDuplicate
}
if err := uc.logtoClient.UpdateUser(ctx, input.LogtoUserID, input.Username,
    input.Name, input.Phone, input.Picture); err != nil {
    return nil, fmt.Errorf("update logto profile: %w", err)
}
```

## 📂 Codebase References

- `internal/app/usecases/profile/setup.go` — `SetupUseCase`, `SetupInput` / `SetupOutput`
- `internal/app/usecases/profile/interfaces.go` — `LogtoUserUpdater` + compile-time check
- `internal/app/ports/user_repo.go` — `UserRepository`
- `internal/domain/errs/errs.go` — `ErrDuplicate`, `ErrInvalidInput`
- `internal/gateways/http/v1/profile/handler.go` — consumer (maps `ErrDuplicate` → 409)

## Verification

```bash
make test
make code-check
go test ./internal/app/usecases/profile/...
```

## Child DOX Index

- No child AGENTS.md files needed — this is a leaf package. Root AGENTS.md owns global standards.
