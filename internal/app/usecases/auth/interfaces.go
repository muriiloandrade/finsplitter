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

// LogtoUserAuthenticator authenticates users via Logto's OIDC ROPC grant.
// Satisfied by *logto.Client in production.
type LogtoUserAuthenticator interface {
	AuthenticateUser(ctx context.Context, email, password string) (*logto.TokenResponse, error)
}
