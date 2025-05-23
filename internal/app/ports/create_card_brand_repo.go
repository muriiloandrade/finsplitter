package ports

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type CreateCardBrandRepository interface {
	CreateCardBrand(ctx context.Context, name string) (entity.CardBrand, error)
}
