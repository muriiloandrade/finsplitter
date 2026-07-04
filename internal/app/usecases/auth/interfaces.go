package auth

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

// Compile-time checks that *logto.Client satisfies our interfaces.
var (
	_ LogtoUserCreator      = (*logto.Client)(nil)
	_ LogtoDeviceFlowClient = (*logto.Client)(nil)
)

// LogtoUserCreator creates users in Logto's identity system.
// Satisfied by *logto.Client in production.
type LogtoUserCreator interface {
	CreateUser(ctx context.Context, username, password, name, email string) (*logto.CreateUserResponse, error)
}

// LogtoDeviceFlowClient initiates and polls the device authorization flow.
// Satisfied by *logto.Client in production.
type LogtoDeviceFlowClient interface {
	RequestDeviceCode(ctx context.Context) (*logto.DeviceCodeResponse, error)
	PollDeviceToken(ctx context.Context, deviceCode string) (*logto.DeviceTokenResponse, error)
}
