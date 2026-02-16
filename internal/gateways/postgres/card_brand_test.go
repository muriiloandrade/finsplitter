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

func TestCardBrandRepository_Create(t *testing.T) {
	db := testutils.NewTestDB(t)
	txManager := &postgres.TxManager{ConnPool: db.Pool()}
	repo := postgres.NewCardBrandRepository(txManager)

	ctx := context.Background()

	t.Run("creates card brand successfully", func(t *testing.T) {
		// When
		result, err := repo.CreateCardBrand(ctx, "Visa")

		// Then
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, "Visa", result.Name)
		assert.NotZero(t, result.CreatedDate)
	})

	t.Run("fails with duplicate name", func(t *testing.T) {
		// Given
		_, err := repo.CreateCardBrand(ctx, "Mastercard")
		require.NoError(t, err)

		// When
		_, err = repo.CreateCardBrand(ctx, "Mastercard")

		// Then
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCardBrandAlreadyExists)
	})
}

func TestCardBrandRepository_GetByID(t *testing.T) {
	db := testutils.NewTestDB(t)
	txManager := &postgres.TxManager{ConnPool: db.Pool()}
	repo := postgres.NewCardBrandRepository(txManager)

	ctx := context.Background()

	t.Run("returns card brand when exists", func(t *testing.T) {
		// Given
		created, err := repo.CreateCardBrand(ctx, "Amex")
		require.NoError(t, err)

		// When
		result, err := repo.GetCardBrandByID(ctx, created.ID)

		// Then
		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, "Amex", result.Name)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		// Given
		fakeID := uuid.Must(uuid.NewV4())

		// When
		_, err := repo.GetCardBrandByID(ctx, fakeID)

		// Then
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCardBrandNotFound)
	})
}

func TestCardBrandRepository_List(t *testing.T) {
	db := testutils.NewTestDB(t)
	txManager := &postgres.TxManager{ConnPool: db.Pool()}
	repo := postgres.NewCardBrandRepository(txManager)

	ctx := context.Background()

	// Seed test data
	brands := []string{"Visa", "Mastercard", "Amex", "Discover"}
	for _, name := range brands {
		_, err := repo.CreateCardBrand(ctx, name)
		require.NoError(t, err)
	}

	t.Run("lists all card brands", func(t *testing.T) {
		// When
		result, err := repo.ListCardBrands(ctx, ports.ListCardBrandFilterOptions{
			PageSize:   10,
			PageNumber: 1,
		})

		// Then
		require.NoError(t, err)
		assert.Len(t, result, 4)
	})

	t.Run("respects pagination", func(t *testing.T) {
		// When
		result, err := repo.ListCardBrands(ctx, ports.ListCardBrandFilterOptions{
			PageSize:   2,
			PageNumber: 1,
		})

		// Then
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("filters by name (case insensitive)", func(t *testing.T) {
		// Given
		name := "Visa"

		// When
		result, err := repo.ListCardBrands(ctx, ports.ListCardBrandFilterOptions{
			Name:       &name,
			PageSize:   10,
			PageNumber: 1,
		})

		// Then
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Visa", result[0].Name)
	})

	t.Run("filters by id", func(t *testing.T) {
		// Given
		created, err := repo.CreateCardBrand(ctx, "Diners")
		require.NoError(t, err)

		// When
		result, err := repo.ListCardBrands(ctx, ports.ListCardBrandFilterOptions{
			ID:         created.ID,
			PageSize:   10,
			PageNumber: 1,
		})

		// Then
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, created.ID, result[0].ID)
	})
}

func TestCardBrandRepository_Update(t *testing.T) {
	db := testutils.NewTestDB(t)
	txManager := &postgres.TxManager{ConnPool: db.Pool()}
	repo := postgres.NewCardBrandRepository(txManager)

	ctx := context.Background()

	t.Run("updates card brand successfully", func(t *testing.T) {
		// Given
		created, err := repo.CreateCardBrand(ctx, "Old Name")
		require.NoError(t, err)

		// When
		result, err := repo.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
			ID:   created.ID,
			Name: "New Name",
		})

		// Then
		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, "New Name", result.Name)
		assert.True(t, result.LastModifiedDate.After(created.LastModifiedDate) || result.LastModifiedDate.Equal(created.LastModifiedDate))
	})

	t.Run("returns error when not found", func(t *testing.T) {
		// Given
		fakeID := uuid.Must(uuid.NewV4())

		// When
		_, err := repo.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
			ID:   fakeID,
			Name: "Test",
		})

		// Then
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCardBrandNotFound)
	})

	t.Run("fails with duplicate name", func(t *testing.T) {
		// Given
		_, err := repo.CreateCardBrand(ctx, "Visa")
		require.NoError(t, err)

		created, err := repo.CreateCardBrand(ctx, "Mastercard")
		require.NoError(t, err)

		// When
		_, err = repo.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
			ID:   created.ID,
			Name: "Visa",
		})

		// Then
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCardBrandAlreadyExists)
	})
}

func TestCardBrandRepository_Delete(t *testing.T) {
	db := testutils.NewTestDB(t)
	txManager := &postgres.TxManager{ConnPool: db.Pool()}
	repo := postgres.NewCardBrandRepository(txManager)

	ctx := context.Background()

	t.Run("deletes card brand successfully", func(t *testing.T) {
		// Given
		created, err := repo.CreateCardBrand(ctx, "ToDelete")
		require.NoError(t, err)

		// When
		result, err := repo.DeleteCardBrand(ctx, created.ID)

		// Then
		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)

		// Verify deleted
		_, err = repo.GetCardBrandByID(ctx, created.ID)
		assert.ErrorIs(t, err, errs.ErrCardBrandNotFound)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		// Given
		fakeID := uuid.Must(uuid.NewV4())

		// When
		_, err := repo.DeleteCardBrand(ctx, fakeID)

		// Then
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCardBrandNotFound)
	})
}
