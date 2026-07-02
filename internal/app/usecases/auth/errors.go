package auth

import "errors"

var (
	// ErrUsernameTaken is returned when the username is already in use in Logto.
	ErrUsernameTaken = errors.New("username already taken")

	// ErrUserAlreadyExists is returned when the user already has a local record.
	ErrUserAlreadyExists = errors.New("user already exists locally")

	// ErrSignInInvalidCredentials is returned when the email or password is
	// invalid during sign-in.
	ErrSignInInvalidCredentials = errors.New("invalid email or password")
)
