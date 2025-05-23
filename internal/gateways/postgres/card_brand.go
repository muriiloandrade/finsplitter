package postgres

import (
	"context"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/muriiloandrade/finsplitter/internal/domain"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/sqlc"
	slogctx "github.com/veqryn/slog-context"
)

const listCardBrandsOp = "postgres.CardBrandRepository.ListCardBrands"

type CardBrandRepository struct {
	domain.Transactioner

	DB *sqlc.Queries
}

func NewCardBrandRepository(db *TxManager) *CardBrandRepository {
	return &CardBrandRepository{
		DB:            sqlc.New(db),
		Transactioner: db,
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

func (r *CardBrandRepository) CreateCardBrand(ctx context.Context, name string) (entity.CardBrand, error) {
	logger := slogctx.FromCtx(ctx)
	brand, err := r.DB.CreateCardBrand(ctx, sqlc.CreateCardBrandParams{Name: name})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return entity.CardBrand{}, domain.NewConflictError("card brand name already exists")
			}
		}
		logger.Error("Failed to create card brand", slog.String("operation", "CreateCardBrand"), slog.Any("error", err))
		return entity.CardBrand{}, err
	}
	return parseToCardBrand(brand), nil
}

func (r *CardBrandRepository) UpdateCardBrand(ctx context.Context, id uuid.UUID, name string) (entity.CardBrand, error) {
	logger := slogctx.FromCtx(ctx)
	row := r.DB.db.QueryRow(ctx, "UPDATE card_brand SET name = $1 WHERE id = $2 RETURNING id, name, created_date, last_modified_date", name, id)
	var brand sqlc.CardBrand
	err := row.Scan(&brand.ID, &brand.Name, &brand.CreatedDate, &brand.LastModifiedDate)
	if err != nil {
		if err == pgx.ErrNoRows {
			return entity.CardBrand{}, pgx.ErrNoRows
		}
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return entity.CardBrand{}, err
			}
		}
		logger.Error("Failed to update card brand", slog.String("operation", "UpdateCardBrand"), slog.Any("error", err))
		return entity.CardBrand{}, err
	}
	return parseToCardBrand(brand), nil
}

func (r *CardBrandRepository) DeleteCardBrand(ctx context.Context, id uuid.UUID) error {
	logger := slogctx.FromCtx(ctx)
	cmd, err := r.DB.db.Exec(ctx, "DELETE FROM card_brand WHERE id = $1", id)
	if err != nil {
		logger.Error("Failed to delete card brand", slog.String("operation", "DeleteCardBrand"), slog.Any("error", err))
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func parseToCardBrand(brand sqlc.CardBrand) entity.CardBrand {
	return entity.CardBrand{
		ID:               brand.ID,
		Name:             brand.Name,
		CreatedDate:      brand.CreatedDate.Time,
		LastModifiedDate: brand.LastModifiedDate.Time,
	}
}
