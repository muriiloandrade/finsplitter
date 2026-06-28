package profile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases/profile"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSetupUseCase_Execute_Success(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)

	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(false, nil)
	userRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(nil).
		Once()

	uc := profile.NewSetupUseCase(userRepo)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "john",
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "john", output.Username)
	assert.NotEmpty(t, output.UserID)
}

func TestSetupUseCase_Execute_EmptyUsername(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)

	uc := profile.NewSetupUseCase(userRepo)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrInvalidInput)
	require.Nil(t, output)
}

func TestSetupUseCase_Execute_AlreadyExists(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)

	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(true, nil)

	uc := profile.NewSetupUseCase(userRepo)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "john",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrDuplicate)
	require.Nil(t, output)
}

func TestSetupUseCase_Execute_ExistsError(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)

	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(false, errors.New("db unavailable"))

	uc := profile.NewSetupUseCase(userRepo)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "john",
	})

	require.Error(t, err)
	require.Nil(t, output)
}

func TestSetupUseCase_Execute_CreateDuplicate(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)

	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(false, nil)
	userRepo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.LogtoUserID == "logto_user_1" && u.Username == "john"
		})).
		Return(errs.ErrDuplicate).
		Once()

	uc := profile.NewSetupUseCase(userRepo)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "john",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrDuplicate)
	require.Nil(t, output)
}

func TestSetupUseCase_Execute_CreateError(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)

	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(false, nil)
	userRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(errors.New("db unavailable")).
		Once()

	uc := profile.NewSetupUseCase(userRepo)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "john",
	})

	require.Error(t, err)
	require.Nil(t, output)
}
