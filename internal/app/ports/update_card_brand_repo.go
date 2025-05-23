package ports

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type UpdateCardBrandRepository interface {
	UpdateCardBrand(ctx context.Context, id uuid.UUID, name string) (entity.CardBrand, error)
}
