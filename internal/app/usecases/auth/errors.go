package auth

import "errors"

var (
	// ErrUsernameTaken is returned when the username is already in use in Logto.
	ErrUsernameTaken = errors.New("username already taken")

	// ErrUserAlreadyExists is returned when the user already has a local record.
	ErrUserAlreadyExists = errors.New("user already exists locally")
)
