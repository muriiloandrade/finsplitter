package errs

import (
	"errors"
	"testing"
)

func TestErrCardBrandNotFound(t *testing.T) {
	err := ErrCardBrandNotFound
	if err == nil {
		t.Fatal("expected error not to be nil")
	}
	if err.Error() != "card brand not found" {
		t.Errorf("expected 'card brand not found', got '%s'", err.Error())
	}
}

func TestErrCardBrandAlreadyExists(t *testing.T) {
	err := ErrCardBrandAlreadyExists
	if err == nil {
		t.Fatal("expected error not to be nil")
	}
	if err.Error() != "card brand already exists" {
		t.Errorf("expected 'card brand already exists', got '%s'", err.Error())
	}
}

func TestErrCardBrandForeignKeyViolation(t *testing.T) {
	err := ErrCardBrandForeignKeyViolation
	if err == nil {
		t.Fatal("expected error not to be nil")
	}
	if err.Error() != "card brand foreign key violation" {
		t.Errorf("expected 'card brand foreign key violation', got '%s'", err.Error())
	}
}

func TestErrDatabaseGeneric(t *testing.T) {
	err := ErrDatabaseGeneric
	if err == nil {
		t.Fatal("expected error not to be nil")
	}
	if err.Error() != "database operation failed" {
		t.Errorf("expected 'database operation failed', got '%s'", err.Error())
	}
}

func TestErrorsIs(t *testing.T) {
	// Sentinel errors should match themselves
	if !errors.Is(ErrCardBrandNotFound, ErrCardBrandNotFound) {
		t.Error("errors.Is should return true for same error")
	}

	// Different sentinel errors should not match
	if errors.Is(ErrCardBrandNotFound, ErrCardBrandAlreadyExists) {
		t.Error("errors.Is should return false for different sentinel errors")
	}
}
