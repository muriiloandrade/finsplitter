package auth_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	auth "github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMeUseCase_Execute_Success_NeedsSetup(t *testing.T) {
	logtoUserID := "logto_user_abc"
	email := "user@example.com"

	repo := ports.NewMockUserRepository(t)
	repo.EXPECT().
		GetByLogtoUserID(mock.Anything, logtoUserID).
		Return(nil, errs.ErrNotFound)

	uc := auth.NewMeUseCase(repo)
	output, err := uc.Execute(context.Background(), auth.MeInput{
		LogtoUserID: logtoUserID,
		Email:       email,
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.True(t, output.NeedsSetup)
	assert.Equal(t, email, output.Email)
	assert.Empty(t, output.ID)
	assert.Empty(t, output.Username)
}

func TestMeUseCase_Execute_Success_AlreadySetup(t *testing.T) {
	logtoUserID := "logto_user_abc"
	email := "user@example.com"
	userID := uuid.Must(uuid.NewV4())

	repo := ports.NewMockUserRepository(t)
	repo.EXPECT().
		GetByLogtoUserID(mock.Anything, logtoUserID).
		Return(&entity.User{
			ID:          userID,
			LogtoUserID: logtoUserID,
			Username:    "john",
			Email:       email,
		}, nil)

	uc := auth.NewMeUseCase(repo)
	output, err := uc.Execute(context.Background(), auth.MeInput{
		LogtoUserID: logtoUserID,
		Email:       email,
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.False(t, output.NeedsSetup)
	assert.Equal(t, userID.String(), output.ID)
	assert.Equal(t, "john", output.Username)
	assert.Equal(t, email, output.Email)
}

func TestMeUseCase_Execute_DatabaseError(t *testing.T) {
	logtoUserID := "logto_user_abc"
	email := "user@example.com"

	repo := ports.NewMockUserRepository(t)
	repo.EXPECT().
		GetByLogtoUserID(mock.Anything, logtoUserID).
		Return(nil, errs.ErrDatabaseGeneric)

	uc := auth.NewMeUseCase(repo)
	output, err := uc.Execute(context.Background(), auth.MeInput{
		LogtoUserID: logtoUserID,
		Email:       email,
	})

	require.Error(t, err)
	assert.Nil(t, output)
}
