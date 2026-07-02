package profile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid/v5"
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
		Create(mock.Anything, "logto_user_1").
		Return(&entity.User{ID: uuid.Must(uuid.NewV4()), LogtoUserID: "logto_user_1"}, nil).
		Once()

	uc := profile.NewSetupUseCase(userRepo)
	output, err := uc.Execute(context.Background(), "logto_user_1")

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.NotEmpty(t, output.UserID)
}

func TestSetupUseCase_Execute_EmptyLogtoUserID(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)

	uc := profile.NewSetupUseCase(userRepo)
	output, err := uc.Execute(context.Background(), "")

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
	output, err := uc.Execute(context.Background(), "logto_user_1")

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
	output, err := uc.Execute(context.Background(), "logto_user_1")

	require.Error(t, err)
	require.Nil(t, output)
}

func TestSetupUseCase_Execute_CreateDuplicate(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)

	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(false, nil)
	userRepo.EXPECT().
		Create(mock.Anything, "logto_user_1").
		Return(nil, errs.ErrDuplicate).
		Once()

	uc := profile.NewSetupUseCase(userRepo)
	output, err := uc.Execute(context.Background(), "logto_user_1")

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
		Create(mock.Anything, "logto_user_1").
		Return(nil, errors.New("db unavailable")).
		Once()

	uc := profile.NewSetupUseCase(userRepo)
	output, err := uc.Execute(context.Background(), "logto_user_1")

	require.Error(t, err)
	require.Nil(t, output)
}
