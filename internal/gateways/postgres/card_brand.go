package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/sqlc"
	slogctx "github.com/veqryn/slog-context"
)

const (
	createCardBrandOp  = "postgres.CardBrandRepository.CreateCardBrand"
	getCardBrandByIDOp = "postgres.CardBrandRepository.GetCardBrandByID"
	listCardBrandsOp   = "postgres.CardBrandRepository.ListCardBrands"
	updateCardBrandOp  = "postgres.CardBrandRepository.UpdateCardBrand"
	deleteCardBrandOp  = "postgres.CardBrandRepository.DeleteCardBrand"
)

type CardBrandRepository struct {
	db   querier
	sqlc *sqlc.Queries
}

func NewCardBrandRepository(db *TxManager) *CardBrandRepository {
	return &CardBrandRepository{
		db:   db,
		sqlc: sqlc.New(db),
	}
}

func (r *CardBrandRepository) CreateCardBrand(
	ctx context.Context,
	name string,
) (*entity.CardBrand, error) {
	logger := slogctx.FromCtx(ctx)

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
	logger := slogctx.FromCtx(ctx)

	brand, err := r.sqlc.GetCardBrand(ctx, sqlc.GetCardBrandParams{
		ID: id,
	})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to get card brand",
			slog.String("operation", getCardBrandByIDOp),
			slog.Any("error", err),
		)
		return nil, err
	}
	cardBrand := parseToCardBrand(brand)
	return &cardBrand, nil
}

func (r *CardBrandRepository) ListCardBrands(
	ctx context.Context,
	opts ports.ListCardBrandFilterOptions,
) ([]entity.CardBrand, error) {
	logger := slogctx.FromCtx(ctx)

	q := psql.
		Select("id", "name", "created_date", "last_modified_date").
		From("card_brand cb").
		Limit(uint64(opts.PageSize)).
		Offset(uint64((opts.PageNumber - 1) * opts.PageSize))

	if !opts.ID.IsNil() {
		q = q.Where(squirrel.Eq{"cb.id": opts.ID})
	}

	if opts.Name != nil && *opts.Name != "" {
		q = q.Where(squirrel.ILike{"cb.name": "%" + *opts.Name + "%"})
	}

	sql, params, err := q.ToSql()
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to build SQL query for listing card brands",
			slog.String("operation", listCardBrandsOp),
			slog.Any("error", err),
		)
		return nil, err
	}

	rows, err := r.db.Query(ctx, sql, params...)
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to list card brands",
			slog.String("operation", listCardBrandsOp),
			slog.Any("error", err),
		)
		return nil, err
	}
	defer rows.Close()

	cardBrandList := make([]entity.CardBrand, 0, opts.PageSize)
	for rows.Next() {
		var brand sqlc.CardBrand
		if err := rows.Scan(&brand.ID, &brand.Name, &brand.CreatedDate, &brand.LastModifiedDate); err != nil {
			logger.ErrorContext(ctx,
				"Failed to scan card brand",
				slog.String("operation", listCardBrandsOp),
				slog.Any("error", err),
			)
			return nil, err
		}
		cardBrandList = append(cardBrandList, parseToCardBrand(brand))
	}
	return cardBrandList, nil
}

func (r *CardBrandRepository) UpdateCardBrand(
	ctx context.Context,
	opts ports.UpdateCardBrandOptions,
) (*entity.CardBrand, error) {
	logger := slogctx.FromCtx(ctx)

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
	logger := slogctx.FromCtx(ctx)

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
