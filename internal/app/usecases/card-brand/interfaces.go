package usecases

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

// CreateCardBrandUseCase defines the interface for creating a card brand
type CreateCardBrandUseCase interface {
	CreateCardBrand(ctx context.Context, name string) (*entity.CardBrand, error)
}

// GetCardBrandByIDUseCase defines the interface for getting a card brand by ID
type GetCardBrandByIDUseCase interface {
	GetCardBrandByID(ctx context.Context, id uuid.UUID) (*entity.CardBrand, error)
}

// ListCardBrandsUseCase defines the interface for listing all card brands
type ListCardBrandsUseCase interface {
	ListCardBrands(ctx context.Context, opts ports.ListCardBrandFilterOptions) ([]entity.CardBrand, error)
}

// UpdateCardBrandUseCase defines the interface for updating a card brand
type UpdateCardBrandUseCase interface {
	UpdateCardBrand(ctx context.Context, opts ports.UpdateCardBrandOptions) (*entity.CardBrand, error)
}

// DeleteCardBrandUseCase defines the interface for deleting a card brand
type DeleteCardBrandUseCase interface {
	DeleteCardBrand(ctx context.Context, id uuid.UUID) (*entity.CardBrand, error)
}
