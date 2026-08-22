package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/sqlc"
	"github.com/muriiloandrade/finsplitter/pkg/logctx"
)

const (
	createCardBrandOp  = "postgres.CardBrandRepository.CreateCardBrand"
	getCardBrandByIDOp = "postgres.CardBrandRepository.GetCardBrandByID"
	listCardBrandsOp   = "postgres.CardBrandRepository.ListCardBrands"
	updateCardBrandOp  = "postgres.CardBrandRepository.UpdateCardBrand"
	deleteCardBrandOp  = "postgres.CardBrandRepository.DeleteCardBrand"
)

type CardBrandRepository struct {
	sqlc *sqlc.Queries
}

func NewCardBrandRepository(db *TxManager) *CardBrandRepository {
	return &CardBrandRepository{
		sqlc: sqlc.New(db),
	}
}

func (r *CardBrandRepository) CreateCardBrand(
	ctx context.Context,
	name string,
) (*entity.CardBrand, error) {
	logger := logctx.FromCtx(ctx)

	brand, err := r.sqlc.CreateCardBrand(ctx, sqlc.CreateCardBrandParams{
		Name: name,
	})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to create card brand",
			slog.String("operation", createCardBrandOp),
			slog.Any("error", err),
		)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			logger.ErrorContext(ctx,
				"Database operation failed",
				slog.String("operation", createCardBrandOp),
				slog.Any("error", err),
			)

			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return nil, errs.ErrCardBrandAlreadyExists
			case pgerrcode.ForeignKeyViolation:
				return nil, errs.ErrCardBrandForeignKeyViolation
			default:
				return nil, errs.ErrDatabaseGeneric
			}
		}
		return nil, err
	}
	cardBrand := parseToCardBrand(brand)
	return &cardBrand, nil
}

func (r *CardBrandRepository) GetCardBrandByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.CardBrand, error) {
	logger := logctx.FromCtx(ctx)

	brand, err := r.sqlc.GetCardBrand(ctx, sqlc.GetCardBrandParams{
		ID: id,
	})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to get card brand",
			slog.String("operation", getCardBrandByIDOp),
			slog.Any("error", err),
		)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrCardBrandNotFound
		}

		return nil, err
	}
	cardBrand := parseToCardBrand(brand)
	return &cardBrand, nil
}

func (r *CardBrandRepository) ListCardBrands(
	ctx context.Context,
	opts ports.ListCardBrandFilterOptions,
) ([]entity.CardBrand, error) {
	logger := logctx.FromCtx(ctx)

	params := sqlc.ListCardBrandsParams{
		Name:       opts.Name,
		PageOffset: opts.Offset(),
		PageSize:   int64(opts.PageSize),
	}
	if !opts.ID.IsNil() {
		id := opts.ID.String()
		params.ID = &id
	}

	brands, err := r.sqlc.ListCardBrands(ctx, params)
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to list card brands",
			slog.String("operation", listCardBrandsOp),
			slog.Any("error", err),
		)
		return nil, err
	}

	cardBrandList := make([]entity.CardBrand, 0, opts.PageSize)
	for _, brand := range brands {
		cardBrandList = append(cardBrandList, parseToCardBrand(brand))
	}
	return cardBrandList, nil
}

func (r *CardBrandRepository) UpdateCardBrand(
	ctx context.Context,
	opts ports.UpdateCardBrandOptions,
) (*entity.CardBrand, error) {
	logger := logctx.FromCtx(ctx)

	cb, err := r.sqlc.UpdateCardBrand(ctx, sqlc.UpdateCardBrandParams{
		ID:   opts.ID,
		Name: opts.Name,
	})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to update card brand",
			slog.String("operation", updateCardBrandOp),
			slog.Any("error", err),
		)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			logger.ErrorContext(ctx,
				"Database operation failed",
				slog.String("operation", updateCardBrandOp),
				slog.Any("error", err),
			)

			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return nil, errs.ErrCardBrandAlreadyExists
			case pgerrcode.ForeignKeyViolation:
				return nil, errs.ErrCardBrandForeignKeyViolation
			default:
				return nil, errs.ErrDatabaseGeneric
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrCardBrandNotFound
		}
		return nil, err
	}

	cardBrand := parseToCardBrand(cb)
	return &cardBrand, nil
}

func (r *CardBrandRepository) DeleteCardBrand(
	ctx context.Context,
	id uuid.UUID,
) (*entity.CardBrand, error) {
	logger := logctx.FromCtx(ctx)

	cb, err := r.sqlc.DeleteCardBrand(ctx, sqlc.DeleteCardBrandParams{
		ID: id,
	})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to delete card brand",
			slog.String("operation", deleteCardBrandOp),
			slog.Any("error", err),
		)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			logger.ErrorContext(ctx,
				"Database operation failed",
				slog.String("operation", deleteCardBrandOp),
				slog.Any("error", err),
			)

			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return nil, errs.ErrCardBrandForeignKeyViolation
			default:
				return nil, errs.ErrDatabaseGeneric
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrCardBrandNotFound
		}

		return nil, err
	}

	cardBrand := parseToCardBrand(cb)
	return &cardBrand, nil
}

func parseToCardBrand(brand sqlc.CardBrand) entity.CardBrand {
	return entity.CardBrand{
		ID:               brand.ID,
		Name:             brand.Name,
		CreatedDate:      brand.CreatedDate.Time,
		LastModifiedDate: brand.LastModifiedDate.Time,
	}
}
