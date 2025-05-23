package ports

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type DeleteCardBrandRepository interface {
	DeleteCardBrand(ctx context.Context, id uuid.UUID) error
}
