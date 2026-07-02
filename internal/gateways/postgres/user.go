package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/sqlc"
	slogctx "github.com/veqryn/slog-context"
)

const (
	createUserOp       = "postgres.UserRepository.Create"
	getUserByIDOp      = "postgres.UserRepository.GetByID"
	getUserByLogtoIDOp = "postgres.UserRepository.GetByLogtoUserID"
	existsByLogtoIDOp  = "postgres.UserRepository.ExistsByLogtoUserID"
)

type UserRepository struct {
	db   querier
	sqlc *sqlc.Queries
}

func NewUserRepository(db *TxManager) ports.UserRepository {
	return &UserRepository{
		db:   db,
		sqlc: sqlc.New(db),
	}
}

func (r *UserRepository) Create(ctx context.Context, logtoUserID string) (*entity.User, error) {
	logger := slogctx.FromCtx(ctx)

	row, err := r.sqlc.CreateUser(ctx, sqlc.CreateUserParams{
		LogtoUserID: ptr(logtoUserID),
	})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to create user",
			slog.String("operation", createUserOp),
			slog.Any("error", err),
		)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return nil, errs.ErrDuplicate
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return newUserFromRow(row.ID, row.LogtoUserID, row.CreatedDate, row.LastModifiedDate), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	logger := slogctx.FromCtx(ctx)

	row, err := r.sqlc.GetUserByID(ctx, sqlc.GetUserByIDParams{ID: id})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to get user by ID",
			slog.String("operation", getUserByIDOp),
			slog.Any("error", err),
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return newUserFromRow(row.ID, row.LogtoUserID, row.CreatedDate, row.LastModifiedDate), nil
}

func (r *UserRepository) GetByLogtoUserID(ctx context.Context, logtoUserID string) (*entity.User, error) {
	logger := slogctx.FromCtx(ctx)

	row, err := r.sqlc.GetUserByLogtoUserID(ctx, sqlc.GetUserByLogtoUserIDParams{
		LogtoUserID: ptr(logtoUserID),
	})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to get user by Logto user ID",
			slog.String("operation", getUserByLogtoIDOp),
			slog.Any("error", err),
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("get user by logto_user_id: %w", err)
	}

	return newUserFromRow(row.ID, row.LogtoUserID, row.CreatedDate, row.LastModifiedDate), nil
}

func (r *UserRepository) ExistsByLogtoUserID(ctx context.Context, logtoUserID string) (bool, error) {
	logger := slogctx.FromCtx(ctx)

	exists, err := r.sqlc.ExistsByLogtoUserID(ctx, sqlc.ExistsByLogtoUserIDParams{
		LogtoUserID: ptr(logtoUserID),
	})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to check user existence by Logto user ID",
			slog.String("operation", existsByLogtoIDOp),
			slog.Any("error", err),
		)
		return false, fmt.Errorf("exists by logto_user_id: %w", err)
	}
	return exists, nil
}

// newUserFromRow builds a domain User from the fields shared by all sqlc user row types.
func newUserFromRow(id uuid.UUID, logtoUserID *string, createdDate, lastModifiedDate pgtype.Timestamptz) *entity.User {
	user := &entity.User{
		ID:          id,
		CreatedDate: createdDate.Time,
	}
	if lastModifiedDate.Valid {
		user.LastModifiedDate = lastModifiedDate.Time
	}
	if logtoUserID != nil {
		user.LogtoUserID = *logtoUserID
	}
	return user
}

// ptr returns a pointer to the given string.
func ptr(s string) *string {
	return &s
}
