package entity

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// User represents a registered Finsplitter user.
// The LogtoUserID is a pseudo-FK to Logto's user table.
// It is NEVER exposed in API responses.
type User struct {
	ID               uuid.UUID `json:"id"`
	LogtoUserID      string    `json:"-"` // Excluded from JSON serialization
	Username         string    `json:"username"`
	Name             string    `json:"name,omitempty"`
	Email            string    `json:"email,omitempty"`
	PhoneNumber      string    `json:"phoneNumber,omitempty"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}
