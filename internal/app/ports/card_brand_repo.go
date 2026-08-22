package ports

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/domain"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type CreateCardBrandRepository interface {
	CreateCardBrand(ctx context.Context, name string) (*entity.CardBrand, error)
}

type GetCardBrandByIDRepository interface {
	GetCardBrandByID(ctx context.Context, id uuid.UUID) (*entity.CardBrand, error)
}

type ListCardBrandFilterOptions struct {
	// Pagination is the shared page-bound options bundle; reuse it for
	// new list endpoints instead of adding PageSize/PageNumber fields.
	domain.Pagination

	ID   uuid.UUID
	Name *string
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
