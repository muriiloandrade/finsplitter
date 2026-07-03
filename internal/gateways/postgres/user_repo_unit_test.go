package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func makeTimestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func defaultUserRowValues() []any {
	ts := makeTimestamp(time.Now())
	return []any{
		uuid.Must(uuid.NewV4()),
		ptr("test-logto-id"),
		ts,
		ts,
	}
}

func newMockUserRepo(queryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row) *UserRepository {
	if queryRowFn == nil {
		queryRowFn = func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{values: defaultUserRowValues()}
		}
	}
	mq := &mockQuerier{queryRowFn: queryRowFn}
	return &UserRepository{db: mq, sqlc: sqlc.New(mq)}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestUserRepo_Create_Success(t *testing.T) {
	repo := newMockUserRepo(nil)
	ctx := context.Background()

	user, err := repo.Create(ctx, "test-logto-id")

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "test-logto-id", user.LogtoUserID)
	assert.NotZero(t, user.CreatedDate)
}

func TestUserRepo_Create_Duplicate(t *testing.T) {
	repo := newMockUserRepo(func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{err: &pgconn.PgError{Code: pgerrcode.UniqueViolation}}
	})
	ctx := context.Background()

	user, err := repo.Create(ctx, "dup-id")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, errs.ErrDuplicate)
}

func TestUserRepo_Create_GenericError(t *testing.T) {
	repo := newMockUserRepo(func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{err: errors.New("connection refused")}
	})
	ctx := context.Background()

	user, err := repo.Create(ctx, "fail-id")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "create user")
	assert.Contains(t, err.Error(), "connection refused")
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestUserRepo_GetByID_Success(t *testing.T) {
	repo := newMockUserRepo(nil)
	ctx := context.Background()

	user, err := repo.GetByID(ctx, uuid.Must(uuid.NewV4()))

	require.NoError(t, err)
	require.NotNil(t, user)
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	repo := newMockUserRepo(func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{err: pgx.ErrNoRows}
	})
	ctx := context.Background()

	user, err := repo.GetByID(ctx, uuid.Must(uuid.NewV4()))

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestUserRepo_GetByID_GenericError(t *testing.T) {
	repo := newMockUserRepo(func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{err: errors.New("db unavailable")}
	})
	ctx := context.Background()

	user, err := repo.GetByID(ctx, uuid.Must(uuid.NewV4()))

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "get user by id")
	assert.Contains(t, err.Error(), "db unavailable")
}

// ---------------------------------------------------------------------------
// GetByLogtoUserID
// ---------------------------------------------------------------------------

func TestUserRepo_GetByLogtoUserID_Success(t *testing.T) {
	repo := newMockUserRepo(nil)
	ctx := context.Background()

	user, err := repo.GetByLogtoUserID(ctx, "test-logto-id")

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "test-logto-id", user.LogtoUserID)
}

func TestUserRepo_GetByLogtoUserID_NotFound(t *testing.T) {
	repo := newMockUserRepo(func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{err: pgx.ErrNoRows}
	})
	ctx := context.Background()

	user, err := repo.GetByLogtoUserID(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

func TestUserRepo_GetByLogtoUserID_GenericError(t *testing.T) {
	repo := newMockUserRepo(func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{err: errors.New("query timeout")}
	})
	ctx := context.Background()

	user, err := repo.GetByLogtoUserID(ctx, "fail-id")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "get user by logto_user_id")
	assert.Contains(t, err.Error(), "query timeout")
}

// ---------------------------------------------------------------------------
// ExistsByLogtoUserID
// ---------------------------------------------------------------------------

func TestUserRepo_ExistsByLogtoUserID_True(t *testing.T) {
	repo := newMockUserRepo(func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{values: []any{true}}
	})
	ctx := context.Background()

	exists, err := repo.ExistsByLogtoUserID(ctx, "test-id")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepo_ExistsByLogtoUserID_False(t *testing.T) {
	repo := newMockUserRepo(func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{values: []any{false}}
	})
	ctx := context.Background()

	exists, err := repo.ExistsByLogtoUserID(ctx, "nonexistent")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepo_ExistsByLogtoUserID_GenericError(t *testing.T) {
	repo := newMockUserRepo(func(_ context.Context, _ string, _ ...any) pgx.Row {
		return &mockRow{err: errors.New("db error")}
	})
	ctx := context.Background()

	exists, err := repo.ExistsByLogtoUserID(ctx, "fail-id")

	require.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "exists by logto_user_id")
	assert.Contains(t, err.Error(), "db error")
}
