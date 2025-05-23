package entity

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	PhoneNumber      *string   `json:"phoneNumber,omitempty"`
	Username         string    `json:"username"`
	PasswordHash     string    `json:"-"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}
