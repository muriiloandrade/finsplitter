package logto

import "errors"

var (
	ErrM2MUnauthorized        = errors.New("logto m2m unauthorized")
	ErrUserExists             = errors.New("user already exists in logto")
	ErrUserNotFound           = errors.New("user not found in logto")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrAppClientNotConfigured = errors.New("logto app client not configured")
	ErrEmailAlreadyInUse      = errors.New("email already in use in logto")
)

// Device flow errors.
var (
	ErrDeviceCodePending      = errors.New("device code authorization pending")
	ErrDeviceCodeExpired      = errors.New("device code expired")
	ErrDeviceCodeAccessDenied = errors.New("device code access denied")
)
