package ports

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type CreateCardBrandRepository interface {
	CreateCardBrand(ctx context.Context, name string) (*entity.CardBrand, error)
}

type GetCardBrandByIDRepository interface {
	GetCardBrandByID(ctx context.Context, id uuid.UUID) (*entity.CardBrand, error)
}

type ListCardBrandFilterOptions struct {
	ID   uuid.UUID
	Name *string
	// PageSize/PageNumber are uint32 so they convert to the sqlc bigint
	// params (int64) without any overflow risk; the HTTP layer bounds them
	// to 1–100 / ≥ 1 via schema validation.
	PageSize   uint32
	PageNumber uint32
}

type ListCardBrandRepository interface {
	ListCardBrands(
		ctx context.Context,
		filter ListCardBrandFilterOptions,
	) ([]entity.CardBrand, error)
}

type UpdateCardBrandOptions struct {
	ID   uuid.UUID
	Name string
}

type UpdateCardBrandRepository interface {
	UpdateCardBrand(ctx context.Context, opts UpdateCardBrandOptions) (*entity.CardBrand, error)
}

type DeleteCardBrandRepository interface {
	DeleteCardBrand(ctx context.Context, id uuid.UUID) (*entity.CardBrand, error)
}
