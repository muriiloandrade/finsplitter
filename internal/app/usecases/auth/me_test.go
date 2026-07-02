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

func TestMeUseCase_Execute_NeedsSetup(t *testing.T) {
	logtoUserID := "logto_user_abc"

	repo := ports.NewMockUserRepository(t)
	repo.EXPECT().
		GetByLogtoUserID(mock.Anything, logtoUserID).
		Return(nil, errs.ErrNotFound)

	uc := auth.NewMeUseCase(repo)
	output, err := uc.Execute(context.Background(), auth.MeInput{
		LogtoUserID: logtoUserID,
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.True(t, output.NeedsSetup)
	assert.Empty(t, output.ID)
}

func TestMeUseCase_Execute_AlreadySetup(t *testing.T) {
	logtoUserID := "logto_user_abc"
	userID := uuid.Must(uuid.NewV4())

	repo := ports.NewMockUserRepository(t)
	repo.EXPECT().
		GetByLogtoUserID(mock.Anything, logtoUserID).
		Return(&entity.User{
			ID:          userID,
			LogtoUserID: logtoUserID,
		}, nil)

	uc := auth.NewMeUseCase(repo)
	output, err := uc.Execute(context.Background(), auth.MeInput{
		LogtoUserID: logtoUserID,
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.False(t, output.NeedsSetup)
	assert.Equal(t, userID.String(), output.ID)
}

func TestMeUseCase_Execute_DatabaseError(t *testing.T) {
	logtoUserID := "logto_user_abc"

	repo := ports.NewMockUserRepository(t)
	repo.EXPECT().
		GetByLogtoUserID(mock.Anything, logtoUserID).
		Return(nil, errs.ErrDatabaseGeneric)

	uc := auth.NewMeUseCase(repo)
	output, err := uc.Execute(context.Background(), auth.MeInput{
		LogtoUserID: logtoUserID,
	})

	require.Error(t, err)
	assert.Nil(t, output)
}
