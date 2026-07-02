package entity

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// User represents a registered Finsplitter user.
// Profile data (name, email, username, phone_number) lives in Logto only.
// The LogtoUserID is a pseudo-FK to Logto's user table.
// It is NEVER exposed in API responses.
type User struct {
	ID               uuid.UUID `json:"id"`
	LogtoUserID      string    `json:"-"` // Excluded from JSON serialization
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}
