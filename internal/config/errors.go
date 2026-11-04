package config

import (
	"errors"
	"strings"
)

// ErrMissingRequiredEnv is the sentinel error returned when required
// environment variables are missing.
var ErrMissingRequiredEnv = errors.New("missing required environment variables")

// missingEnvError carries the names of the required environment variables that
// are missing, while unwrapping to the ErrMissingRequiredEnv sentinel.
type missingEnvError struct {
	missing []string
}

// Error implements the error interface.
func (e *missingEnvError) Error() string {
	return "missing required environment variables: " + strings.Join(e.missing, ", ")
}

// Unwrap exposes the sentinel error for errors.Is.
func (e *missingEnvError) Unwrap() error {
	return ErrMissingRequiredEnv
}
