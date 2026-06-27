package logto

import "errors"

var (
	ErrM2MUnauthorized = errors.New("logto m2m unauthorized")
	ErrUserExists      = errors.New("user already exists in logto")
)
