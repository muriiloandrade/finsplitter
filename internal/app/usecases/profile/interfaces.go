package profile

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

// Compile-time check that *logto.Client satisfies our interfaces.
var (
	_ LogtoUserUpdater = (*logto.Client)(nil)
)

// LogtoUserUpdater updates user profile fields in Logto via the Management API.
// Satisfied by *logto.Client in production.
type LogtoUserUpdater interface {
	UpdateUser(ctx context.Context, userID, username, name, phone, picture string) error
}
