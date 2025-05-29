package errs

import "errors"

var (
	ErrDatabaseGeneric = errors.New("database operation failed")
)

var (
	ErrCardBrandNotFound            = errors.New("card brand not found")
	ErrCardBrandAlreadyExists       = errors.New("card brand already exists")
	ErrCardBrandForeignKeyViolation = errors.New("card brand foreign key violation")
)
