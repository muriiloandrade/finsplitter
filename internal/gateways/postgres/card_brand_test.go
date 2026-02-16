package postgres_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRepo(t *testing.T) (context.Context, *postgres.CardBrandRepository) {
	db := testutils.NewTestDB(t)
	txManager := &postgres.TxManager{ConnPool: db.Pool()}
	repo := postgres.NewCardBrandRepository(txManager)
	return context.Background(), repo
}

func TestCardBrandRepository_Create(t *testing.T) {
	ctx, repo := setupRepo(t)

	// Setup: create first brand for duplicate test
	_, err := repo.CreateCardBrand(ctx, "Mastercard")
	require.NoError(t, err)

	tests := []struct {
		name         string
		brandName    string
		wantErr      error
		wantName     string
		wantIDNotNil bool
	}{
		{
			name:         "creates card brand successfully",
			brandName:    "Visa",
			wantErr:      nil,
			wantName:     "Visa",
			wantIDNotNil: true,
		},
		{
			name:         "fails with duplicate name",
			brandName:    "Mastercard",
			wantErr:      errs.ErrCardBrandAlreadyExists,
			wantIDNotNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, createErr := repo.CreateCardBrand(ctx, tt.brandName)

			if tt.wantErr != nil {
				require.Error(t, createErr)
				assert.ErrorIs(t, createErr, tt.wantErr)
				return
			}

			require.NoError(t, createErr)
			if tt.wantIDNotNil {
				assert.NotEqual(t, uuid.Nil, result.ID)
			}
			assert.Equal(t, tt.wantName, result.Name)
			assert.NotZero(t, result.CreatedDate)
		})
	}
}

func TestCardBrandRepository_GetByID(t *testing.T) {
	ctx, repo := setupRepo(t)

	tests := []struct {
		name     string
		setupID  func() uuid.UUID
		wantErr  error
		wantName string
	}{
		{
			name: "returns card brand when exists",
			setupID: func() uuid.UUID {
				created, err := repo.CreateCardBrand(ctx, "Amex")
				require.NoError(t, err)
				return created.ID
			},
			wantErr:  nil,
			wantName: "Amex",
		},
		{
			name: "returns error when not found",
			setupID: func() uuid.UUID {
				return uuid.Must(uuid.NewV4())
			},
			wantErr: errs.ErrCardBrandNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.setupID()

			result, err := repo.GetCardBrandByID(ctx, id)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, id, result.ID)
			assert.Equal(t, tt.wantName, result.Name)
		})
	}
}

func TestCardBrandRepository_List(t *testing.T) {
	ctx, repo := setupRepo(t)

	// Seed test data
	brands := []string{"Visa", "Mastercard", "Amex", "Discover"}
	for _, name := range brands {
		_, err := repo.CreateCardBrand(ctx, name)
		require.NoError(t, err)
	}

	// Create an additional brand to get its ID for filtering
	diners, err := repo.CreateCardBrand(ctx, "Diners")
	require.NoError(t, err)

	tests := []struct {
		name      string
		filter    ports.ListCardBrandFilterOptions
		wantCount int
		wantFirst *string
	}{
		{
			name: "lists all card brands",
			filter: ports.ListCardBrandFilterOptions{
				PageSize:   10,
				PageNumber: 1,
			},
			wantCount: 5,
		},
		{
			name: "respects pagination",
			filter: ports.ListCardBrandFilterOptions{
				PageSize:   2,
				PageNumber: 1,
			},
			wantCount: 2,
		},
		{
			name: "filters by name (case insensitive)",
			filter: ports.ListCardBrandFilterOptions{
				Name:       strPtr("Visa"),
				PageSize:   10,
				PageNumber: 1,
			},
			wantCount: 1,
			wantFirst: strPtr("Visa"),
		},
		{
			name: "filters by id",
			filter: ports.ListCardBrandFilterOptions{
				ID:         diners.ID,
				PageSize:   10,
				PageNumber: 1,
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, listErr := repo.ListCardBrands(ctx, tt.filter)

			require.NoError(t, listErr)
			assert.Len(t, result, tt.wantCount)

			if tt.wantFirst != nil && len(result) > 0 {
				assert.Equal(t, *tt.wantFirst, result[0].Name)
			}
		})
	}
}

func TestCardBrandRepository_Update(t *testing.T) {
	ctx, repo := setupRepo(t)

	tests := []struct {
		name       string
		setup      func() (uuid.UUID, error)
		updateID   func(id uuid.UUID) uuid.UUID
		updateName string
		wantErr    error
		wantName   string
	}{
		{
			name: "updates card brand successfully",
			setup: func() (uuid.UUID, error) {
				created, err := repo.CreateCardBrand(ctx, "Old Name")
				return created.ID, err
			},
			updateID:   func(id uuid.UUID) uuid.UUID { return id },
			updateName: "New Name",
			wantErr:    nil,
			wantName:   "New Name",
		},
		{
			name: "returns error when not found",
			setup: func() (uuid.UUID, error) {
				return uuid.Must(uuid.NewV4()), nil
			},
			updateID:   func(id uuid.UUID) uuid.UUID { return id },
			updateName: "Test",
			wantErr:    errs.ErrCardBrandNotFound,
		},
		{
			name: "fails with duplicate name",
			setup: func() (uuid.UUID, error) {
				_, err := repo.CreateCardBrand(ctx, "Visa")
				if err != nil {
					return uuid.Nil, err
				}
				created, err := repo.CreateCardBrand(ctx, "Mastercard")
				return created.ID, err
			},
			updateID:   func(id uuid.UUID) uuid.UUID { return id },
			updateName: "Visa",
			wantErr:    errs.ErrCardBrandAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := tt.setup()
			require.NoError(t, err)

			updateID := tt.updateID(id)
			result, err := repo.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
				ID:   updateID,
				Name: tt.updateName,
			})

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, updateID, result.ID)
			assert.Equal(t, tt.wantName, result.Name)
		})
	}
}

func TestCardBrandRepository_Delete(t *testing.T) {
	ctx, repo := setupRepo(t)

	tests := []struct {
		name    string
		setupID func() uuid.UUID
		wantErr error
	}{
		{
			name: "deletes card brand successfully",
			setupID: func() uuid.UUID {
				created, err := repo.CreateCardBrand(ctx, "ToDelete")
				require.NoError(t, err)
				return created.ID
			},
			wantErr: nil,
		},
		{
			name: "returns error when not found",
			setupID: func() uuid.UUID {
				return uuid.Must(uuid.NewV4())
			},
			wantErr: errs.ErrCardBrandNotFound,
		},
	}

	// Note: ForeignKeyViolation test requires a dependent table with data,
	// which is not set up in this test. The code path exists at line 232.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.setupID()

			result, err := repo.DeleteCardBrand(ctx, id)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, id, result.ID)

			// Verify deleted
			_, err = repo.GetCardBrandByID(ctx, id)
			assert.ErrorIs(t, err, errs.ErrCardBrandNotFound)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
