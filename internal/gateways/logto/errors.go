package logto

import "errors"

var (
	ErrM2MUnauthorized      = errors.New("logto m2m unauthorized")
	ErrUserExists           = errors.New("user already exists in logto")
	ErrUserNotFound         = errors.New("user not found in logto")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrAppClientNotConfigured = errors.New("logto app client not configured")
)
