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
	updateUsernameOp   = "postgres.UserRepository.UpdateUsername"
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

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	logger := slogctx.FromCtx(ctx)

	row, err := r.sqlc.CreateUser(ctx, sqlc.CreateUserParams{
		LogtoUserID: ptr(user.LogtoUserID),
		Username:    user.Username,
		Email:       user.Email,
		Name:        user.Name,
		PhoneNumber: ptrNonEmpty(user.PhoneNumber),
	})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to create user",
			slog.String("operation", createUserOp),
			slog.Any("error", err),
		)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return errs.ErrDuplicate
		}
		return fmt.Errorf("create user: %w", err)
	}

	user.ID = row.ID
	user.CreatedDate = row.CreatedDate.Time
	user.LastModifiedDate = row.LastModifiedDate.Time
	if row.Email != "" {
		user.Email = row.Email
	}
	return nil
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

	return mapUserRow(userRow{
		ID:               row.ID,
		Username:         row.Username,
		Email:            row.Email,
		LogtoUserID:      row.LogtoUserID,
		CreatedDate:      row.CreatedDate,
		LastModifiedDate: row.LastModifiedDate,
	}), nil
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

	return mapUserRow(userRow{
		ID:               row.ID,
		Username:         row.Username,
		Email:            row.Email,
		LogtoUserID:      row.LogtoUserID,
		CreatedDate:      row.CreatedDate,
		LastModifiedDate: row.LastModifiedDate,
	}), nil
}

func (r *UserRepository) UpdateUsername(ctx context.Context, id uuid.UUID, username string) error {
	logger := slogctx.FromCtx(ctx)

	err := r.sqlc.UpdateUsername(ctx, sqlc.UpdateUsernameParams{
		ID:       id,
		Username: username,
	})
	if err != nil {
		logger.ErrorContext(ctx,
			"Failed to update username",
			slog.String("operation", updateUsernameOp),
			slog.Any("error", err),
		)
		return fmt.Errorf("update username: %w", err)
	}
	return nil
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

// FindUsernamesByPrefix returns usernames that start with the given LIKE pattern.
func (r *UserRepository) FindUsernamesByPrefix(ctx context.Context, prefix string) ([]string, error) {
	return r.sqlc.FindUsernamesByPrefix(ctx, sqlc.FindUsernamesByPrefixParams{Username: prefix})
}

// userRow is an internal aggregate that carries the fields needed to build a
// domain User from any sqlc row type (GetUserByIDRow, GetUserByLogtoUserIDRow,
// CreateUserRow — all share the same shape).
type userRow struct {
	ID               uuid.UUID
	Username         string
	Email            string
	LogtoUserID      *string
	CreatedDate      pgtype.Timestamptz
	LastModifiedDate pgtype.Timestamptz
}

// mapUserRow maps a userRow aggregate to the domain User entity.
func mapUserRow(row userRow) *entity.User {
	user := &entity.User{
		ID:               row.ID,
		Username:         row.Username,
		Email:            row.Email,
		CreatedDate:      row.CreatedDate.Time,
		LastModifiedDate: row.LastModifiedDate.Time,
	}
	if row.LogtoUserID != nil {
		user.LogtoUserID = *row.LogtoUserID
	}
	return user
}

// ptr returns a pointer to the given string.
func ptr(s string) *string {
	return &s
}

// ptrNonEmpty returns a pointer to s when s is non-empty, or nil otherwise.
func ptrNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
