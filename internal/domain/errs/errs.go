package errs

import "errors"

var ErrDatabaseGeneric = errors.New("database operation failed")

var (
	ErrCardBrandNotFound            = errors.New("card brand not found")
	ErrCardBrandAlreadyExists       = errors.New("card brand already exists")
	ErrCardBrandForeignKeyViolation = errors.New("card brand foreign key violation")
)

// Auth/User domain errors.
var (
	ErrNotFound           = errors.New("entity not found")
	ErrDuplicate          = errors.New("duplicate entry")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrNeedsSetup         = errors.New("account needs profile setup")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUsernameTaken      = errors.New("username already taken")
)
