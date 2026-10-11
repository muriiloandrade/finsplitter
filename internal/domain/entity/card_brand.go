package entity

import (
	"time"
	"uuid"
)

type CardBrand struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}
