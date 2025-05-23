package usecases

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
)

type DeleteCardBrandUC interface {
	DeleteCardBrand(ctx context.Context, id uuid.UUID) error
}

type DeleteCardBrandUseCase struct {
	repo ports.DeleteCardBrandRepository
}

func NewDeleteCardBrandUC(repo ports.DeleteCardBrandRepository) *DeleteCardBrandUseCase {
	return &DeleteCardBrandUseCase{repo: repo}
}

func (uc *DeleteCardBrandUseCase) DeleteCardBrand(ctx context.Context, id uuid.UUID) error {
	return uc.repo.DeleteCardBrand(ctx, id)
}
