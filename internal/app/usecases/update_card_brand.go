package usecases

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type UpdateCardBrandUC interface {
	UpdateCardBrand(ctx context.Context, id uuid.UUID, name string) (entity.CardBrand, error)
}

type UpdateCardBrandUseCase struct {
	repo ports.UpdateCardBrandRepository
}

func NewUpdateCardBrandUC(repo ports.UpdateCardBrandRepository) *UpdateCardBrandUseCase {
	return &UpdateCardBrandUseCase{repo: repo}
}

func (uc *UpdateCardBrandUseCase) UpdateCardBrand(ctx context.Context, id uuid.UUID, name string) (entity.CardBrand, error) {
	return uc.repo.UpdateCardBrand(ctx, id, name)
}
