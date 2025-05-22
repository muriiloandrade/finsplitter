package entity

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CardBrand struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}
