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

const testLogtoUserID = "test-logto-user-1"

func setupUserRepo(t *testing.T) (context.Context, ports.UserRepository) {
	t.Helper()

	db := testutils.NewTestDB(t)
	txManager := &postgres.TxManager{ConnPool: db.Pool()}
	repo := postgres.NewUserRepository(txManager)
	return context.Background(), repo
}

// createTestUser inserts a user for tests that need an existing record.
func createTestUser(ctx context.Context, t *testing.T, repo ports.UserRepository, logtoUserID string) uuid.UUID {
	t.Helper()

	user, err := repo.Create(ctx, logtoUserID)
	require.NoError(t, err)
	require.NotNil(t, user)
	return user.ID
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestUserRepository_Create_Success(t *testing.T) {
	ctx, repo := setupUserRepo(t)

	user, err := repo.Create(ctx, testLogtoUserID)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, testLogtoUserID, user.LogtoUserID)
	assert.NotZero(t, user.CreatedDate)
	assert.NotZero(t, user.LastModifiedDate)
}

func TestUserRepository_Create_Duplicate(t *testing.T) {
	ctx, repo := setupUserRepo(t)

	// Create first user
	_, err := repo.Create(ctx, "duplicate-id")
	require.NoError(t, err)

	// Attempt duplicate
	user, err := repo.Create(ctx, "duplicate-id")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, errs.ErrDuplicate)
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestUserRepository_GetByID_Success(t *testing.T) {
	ctx, repo := setupUserRepo(t)

	createdID := createTestUser(ctx, t, repo, testLogtoUserID)

	user, err := repo.GetByID(ctx, createdID)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, createdID, user.ID)
	assert.Equal(t, testLogtoUserID, user.LogtoUserID)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	ctx, repo := setupUserRepo(t)

	user, err := repo.GetByID(ctx, uuid.Must(uuid.NewV4()))

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

// ---------------------------------------------------------------------------
// GetByLogtoUserID
// ---------------------------------------------------------------------------

func TestUserRepository_GetByLogtoUserID_Success(t *testing.T) {
	ctx, repo := setupUserRepo(t)

	_ = createTestUser(ctx, t, repo, testLogtoUserID)

	user, err := repo.GetByLogtoUserID(ctx, testLogtoUserID)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, testLogtoUserID, user.LogtoUserID)
}

func TestUserRepository_GetByLogtoUserID_NotFound(t *testing.T) {
	ctx, repo := setupUserRepo(t)

	user, err := repo.GetByLogtoUserID(ctx, "nonexistent-logto-id")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

// ---------------------------------------------------------------------------
// ExistsByLogtoUserID
// ---------------------------------------------------------------------------

func TestUserRepository_ExistsByLogtoUserID_True(t *testing.T) {
	ctx, repo := setupUserRepo(t)

	_ = createTestUser(ctx, t, repo, testLogtoUserID)

	exists, err := repo.ExistsByLogtoUserID(ctx, testLogtoUserID)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_ExistsByLogtoUserID_False(t *testing.T) {
	ctx, repo := setupUserRepo(t)

	exists, err := repo.ExistsByLogtoUserID(ctx, "nonexistent-id")

	require.NoError(t, err)
	assert.False(t, exists)
}
