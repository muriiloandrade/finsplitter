package usecases

import (
	"context"
	"log/slog"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	slogctx "github.com/veqryn/slog-context"
)

const operation = "usecases.ListCardBrands"

type ListCardBrandsUC struct {
	repo ports.ListCardBrandRepository
}

func NewListCardBrandUC(repo ports.ListCardBrandRepository) ListCardBrandsUC {
	return ListCardBrandsUC{repo: repo}
}

func (uc *ListCardBrandsUC) ListCardBrands(ctx context.Context, opts ports.ListCardBrandFilterOptions) ([]entity.CardBrand, error) {
	logger := slogctx.FromCtx(ctx)

	cardBrands, err := uc.repo.ListCardBrands(ctx, opts)
	if err != nil {
		logger.Error("Failed to list card brands", slog.String("operation", operation), slog.Any("error", err))
		return nil, err
	}

	return cardBrands, nil
}
