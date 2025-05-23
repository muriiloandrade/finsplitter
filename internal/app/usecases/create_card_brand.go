package usecases

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type CreateCardBrandUC interface {
	CreateCardBrand(ctx context.Context, name string) (entity.CardBrand, error)
}

type CreateCardBrandUseCase struct {
	repo ports.CreateCardBrandRepository
}

func NewCreateCardBrandUC(repo ports.CreateCardBrandRepository) *CreateCardBrandUseCase {
	return &CreateCardBrandUseCase{repo: repo}
}

func (uc *CreateCardBrandUseCase) CreateCardBrand(ctx context.Context, name string) (entity.CardBrand, error) {
	return uc.repo.CreateCardBrand(ctx, name)
}
