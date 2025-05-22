package postgres

import (
	"context"
	"log/slog"

	"github.com/muriiloandrade/finsplitter/app/domain/entity"
	"github.com/muriiloandrade/finsplitter/app/gateways/postgres/sqlc"
	slogctx "github.com/veqryn/slog-context"
)

const listCardBrandsOp = "postgres.CardBrandRepository.ListCardBrands"

type CardBrandRepository struct {
	DB *sqlc.Queries
}

func NewCardBrandRepository(db *TxManager) *CardBrandRepository {
	return &CardBrandRepository{
		DB: sqlc.New(db),
	}
}

func (r *CardBrandRepository) ListCardBrands(ctx context.Context) ([]entity.CardBrand, error) {
	logger := slogctx.FromCtx(ctx)

	brands, err := r.DB.ListCardBrands(ctx)
	if err != nil {
		logger.Error("Failed to list card brands", slog.String("operation", listCardBrandsOp), slog.Any("error", err))
		return nil, err
	}

	cardBrandList := make([]entity.CardBrand, 0, len(brands))
	for _, brand := range brands {
		cardBrandList = append(cardBrandList, parseToCardBrand(brand))
	}
	return cardBrandList, nil
}

func parseToCardBrand(brand sqlc.CardBrand) entity.CardBrand {
	return entity.CardBrand{
		ID:               brand.ID,
		Name:             brand.Name,
		CreatedDate:      brand.CreatedDate.Time,
		LastModifiedDate: brand.LastModifiedDate.Time,
	}
}
