package auth

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

// LogtoUserCreator creates users in Logto's identity system.
// Satisfied by *logto.Client in production.
type LogtoUserCreator interface {
	CreateUser(ctx context.Context, username, password, name, email string) (*logto.CreateUserResponse, error)
}
