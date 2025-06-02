package usecases

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

const operation = "usecases.ListCardBrands"

type ListCardBrandsUC struct {
	repo ports.ListCardBrandRepository
}

func NewListCardBrandUC(repo ports.ListCardBrandRepository) ListCardBrandsUC {
	return ListCardBrandsUC{repo: repo}
}

func (uc *ListCardBrandsUC) ListCardBrands(ctx context.Context, opts ports.ListCardBrandFilterOptions) ([]entity.CardBrand, error) {
	cardBrands, err := uc.repo.ListCardBrands(ctx, opts)
	if err != nil {
		return nil, err
	}

	return cardBrands, nil
}
