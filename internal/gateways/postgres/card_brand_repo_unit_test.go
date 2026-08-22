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
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

type mockCBFuncs struct {
	queryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
	queryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func newMockCardBrandRepo(fns mockCBFuncs) *CardBrandRepository {
	mq := &mockQuerier{
		queryRowFn: fns.queryRowFn,
		queryFunc:  fns.queryFn,
	}
	return &CardBrandRepository{sqlc: sqlc.New(mq)}
}

func defaultCBRow() *mockRow {
	ts := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	return &mockRow{
		values: []any{uuid.Must(uuid.NewV4()), "Visa", ts, ts},
	}
}

// ---------------------------------------------------------------------------
// CreateCardBrand
// ---------------------------------------------------------------------------

func TestCardBrandRepo_Create_Success(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return defaultCBRow()
		},
	})
	ctx := context.Background()

	brand, err := repo.CreateCardBrand(ctx, "Visa")

	require.NoError(t, err)
	require.NotNil(t, brand)
	assert.Equal(t, "Visa", brand.Name)
}

func TestCardBrandRepo_Create_UniqueViolation(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: &pgconn.PgError{Code: pgerrcode.UniqueViolation}}
		},
	})
	ctx := context.Background()

	brand, err := repo.CreateCardBrand(ctx, "Duplicate")

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrCardBrandAlreadyExists)
}

func TestCardBrandRepo_Create_ForeignKeyViolation(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation}}
		},
	})
	ctx := context.Background()

	brand, err := repo.CreateCardBrand(ctx, "Invalid")

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrCardBrandForeignKeyViolation)
}

func TestCardBrandRepo_Create_DefaultPgError(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: &pgconn.PgError{Code: pgerrcode.CheckViolation}}
		},
	})
	ctx := context.Background()

	brand, err := repo.CreateCardBrand(ctx, "Bad")

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrDatabaseGeneric)
}

func TestCardBrandRepo_Create_NonPgError(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: errors.New("network timeout")}
		},
	})
	ctx := context.Background()

	brand, err := repo.CreateCardBrand(ctx, "Fail")

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.Contains(t, err.Error(), "network timeout")
}

// ---------------------------------------------------------------------------
// GetCardBrandByID
// ---------------------------------------------------------------------------

func TestCardBrandRepo_GetByID_Success(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return defaultCBRow()
		},
	})
	ctx := context.Background()

	brand, err := repo.GetCardBrandByID(ctx, uuid.Must(uuid.NewV4()))

	require.NoError(t, err)
	require.NotNil(t, brand)
	assert.Equal(t, "Visa", brand.Name)
}

func TestCardBrandRepo_GetByID_NotFound(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: pgx.ErrNoRows}
		},
	})
	ctx := context.Background()

	brand, err := repo.GetCardBrandByID(ctx, uuid.Must(uuid.NewV4()))

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrCardBrandNotFound)
}

func TestCardBrandRepo_GetByID_NonPgError(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: errors.New("db error")}
		},
	})
	ctx := context.Background()

	brand, err := repo.GetCardBrandByID(ctx, uuid.Must(uuid.NewV4()))

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.Contains(t, err.Error(), "db error")
}

// ---------------------------------------------------------------------------
// UpdateCardBrand
// ---------------------------------------------------------------------------

func TestCardBrandRepo_Update_Success(t *testing.T) {
	ts := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{
				values: []any{uuid.Must(uuid.NewV4()), "UpdatedName", ts, ts},
			}
		},
	})
	ctx := context.Background()

	brand, err := repo.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
		ID:   uuid.Must(uuid.NewV4()),
		Name: "UpdatedName",
	})

	require.NoError(t, err)
	require.NotNil(t, brand)
	assert.Equal(t, "UpdatedName", brand.Name)
}

func TestCardBrandRepo_Update_NotFound(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: pgx.ErrNoRows}
		},
	})
	ctx := context.Background()

	brand, err := repo.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
		ID:   uuid.Must(uuid.NewV4()),
		Name: "Ghost",
	})

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrCardBrandNotFound)
}

func TestCardBrandRepo_Update_UniqueViolation(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: &pgconn.PgError{Code: pgerrcode.UniqueViolation}}
		},
	})
	ctx := context.Background()

	brand, err := repo.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
		ID:   uuid.Must(uuid.NewV4()),
		Name: "Duplicate",
	})

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrCardBrandAlreadyExists)
}

func TestCardBrandRepo_Update_ForeignKeyViolation(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation}}
		},
	})
	ctx := context.Background()

	brand, err := repo.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
		ID:   uuid.Must(uuid.NewV4()),
		Name: "Invalid",
	})

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrCardBrandForeignKeyViolation)
}

func TestCardBrandRepo_Update_DefaultPgError(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: &pgconn.PgError{Code: pgerrcode.CheckViolation}}
		},
	})
	ctx := context.Background()

	brand, err := repo.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
		ID:   uuid.Must(uuid.NewV4()),
		Name: "Bad",
	})

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrDatabaseGeneric)
}

func TestCardBrandRepo_Update_NonPgError(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: errors.New("timeout")}
		},
	})
	ctx := context.Background()

	brand, err := repo.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
		ID:   uuid.Must(uuid.NewV4()),
		Name: "Fail",
	})

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.Contains(t, err.Error(), "timeout")
}

// ---------------------------------------------------------------------------
// DeleteCardBrand
// ---------------------------------------------------------------------------

func TestCardBrandRepo_Delete_Success(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return defaultCBRow()
		},
	})
	ctx := context.Background()

	brand, err := repo.DeleteCardBrand(ctx, uuid.Must(uuid.NewV4()))

	require.NoError(t, err)
	require.NotNil(t, brand)
}

func TestCardBrandRepo_Delete_NotFound(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: pgx.ErrNoRows}
		},
	})
	ctx := context.Background()

	brand, err := repo.DeleteCardBrand(ctx, uuid.Must(uuid.NewV4()))

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrCardBrandNotFound)
}

func TestCardBrandRepo_Delete_ForeignKeyViolation(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation}}
		},
	})
	ctx := context.Background()

	brand, err := repo.DeleteCardBrand(ctx, uuid.Must(uuid.NewV4()))

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrCardBrandForeignKeyViolation)
}

func TestCardBrandRepo_Delete_DefaultPgError(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: &pgconn.PgError{Code: pgerrcode.CheckViolation}}
		},
	})
	ctx := context.Background()

	brand, err := repo.DeleteCardBrand(ctx, uuid.Must(uuid.NewV4()))

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.ErrorIs(t, err, errs.ErrDatabaseGeneric)
}

func TestCardBrandRepo_Delete_NonPgError(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{err: errors.New("unexpected error")}
		},
	})
	ctx := context.Background()

	brand, err := repo.DeleteCardBrand(ctx, uuid.Must(uuid.NewV4()))

	require.Error(t, err)
	assert.Nil(t, brand)
	assert.Contains(t, err.Error(), "unexpected error")
}

// ---------------------------------------------------------------------------
// ListCardBrands
// ---------------------------------------------------------------------------

func TestCardBrandRepo_List_Success(t *testing.T) {
	ts := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{
				rows: [][]any{
					{uuid.Must(uuid.NewV4()), "Visa", ts, ts},
					{uuid.Must(uuid.NewV4()), "Mastercard", ts, ts},
				},
			}, nil
		},
	})
	ctx := context.Background()

	brands, err := repo.ListCardBrands(ctx, ports.ListCardBrandFilterOptions{
		PageSize:   10,
		PageNumber: 1,
	})

	require.NoError(t, err)
	assert.Len(t, brands, 2)
	assert.Equal(t, "Visa", brands[0].Name)
	assert.Equal(t, "Mastercard", brands[1].Name)
}

func TestCardBrandRepo_List_Empty(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{}, nil
		},
	})
	ctx := context.Background()

	brands, err := repo.ListCardBrands(ctx, ports.ListCardBrandFilterOptions{
		PageSize:   10,
		PageNumber: 1,
	})

	require.NoError(t, err)
	assert.Empty(t, brands)
}

func TestCardBrandRepo_List_QueryError(t *testing.T) {
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return nil, errors.New("query execution failed")
		},
	})
	ctx := context.Background()

	brands, err := repo.ListCardBrands(ctx, ports.ListCardBrandFilterOptions{
		PageSize:   10,
		PageNumber: 1,
	})

	require.Error(t, err)
	assert.Nil(t, brands)
	assert.Contains(t, err.Error(), "query execution failed")
}

func TestCardBrandRepo_List_ScanError(t *testing.T) {
	ts := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	repo := newMockCardBrandRepo(mockCBFuncs{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{
				rows:    [][]any{{uuid.Must(uuid.NewV4()), "Visa", ts, ts}},
				scanErr: errors.New("column mismatch"),
			}, nil
		},
	})
	ctx := context.Background()

	brands, err := repo.ListCardBrands(ctx, ports.ListCardBrandFilterOptions{
		PageSize:   10,
		PageNumber: 1,
	})

	require.Error(t, err)
	assert.Nil(t, brands)
	assert.Contains(t, err.Error(), "column mismatch")
}
