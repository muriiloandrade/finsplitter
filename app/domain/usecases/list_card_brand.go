package usecases

import (
	"context"
	"log/slog"

	"github.com/muriiloandrade/finsplitter/app/domain/entity"
	slogctx "github.com/veqryn/slog-context"
)

const operation = "usecases.ListCardBrands"

type ListCardBrandUC interface {
	ListCardBrands(ctx context.Context) ([]entity.CardBrand, error)
}

type ListCardBrandRepository interface {
	ListCardBrands(ctx context.Context) ([]entity.CardBrand, error)
}

type ListCardBrandsUseCase struct {
	repo ListCardBrandRepository
}

func NewListCardBrandUC(repo ListCardBrandRepository) *ListCardBrandsUseCase {
	return &ListCardBrandsUseCase{
		repo: repo,
	}
}

func (uc *ListCardBrandsUseCase) ListCardBrands(ctx context.Context) ([]entity.CardBrand, error) {
	logger := slogctx.FromCtx(ctx)

	cardBrands, err := uc.repo.ListCardBrands(ctx)
	if err != nil {
		logger.Error("Failed to list card brands", slog.String("operation", operation), slog.Any("error", err))
		return nil, err
	}

	return cardBrands, nil
}
