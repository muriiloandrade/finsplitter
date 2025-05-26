package ports

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type ListCardBrandRepository interface {
	ListCardBrands(ctx context.Context) ([]entity.CardBrand, error)
}
