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

// mockLogtoUpdater is a manual mock for profile.LogtoUserUpdater.
type mockLogtoUpdater struct {
	mock.Mock
}

func (m *mockLogtoUpdater) UpdateUser(
	ctx context.Context, userID, username, name, phone, picture string,
) error {
	args := m.Called(ctx, userID, username, name, phone, picture)
	return args.Error(0)
}

func TestSetupUseCase_Execute_Success(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)
	logtoClient := new(mockLogtoUpdater)

	logtoClient.On("UpdateUser", mock.Anything, "logto_user_1", "john", "John Doe", "+551199999999", "https://example.com/avatar.jpg").
		Return(nil)
	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(false, nil)
	userRepo.EXPECT().
		Create(mock.Anything, "logto_user_1").
		Return(&entity.User{ID: uuid.Must(uuid.NewV4()), LogtoUserID: "logto_user_1"}, nil).
		Once()

	uc := profile.NewSetupUseCase(userRepo, logtoClient)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "john",
		Name:        "John Doe",
		Phone:       "+551199999999",
		Picture:     "https://example.com/avatar.jpg",
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.NotEmpty(t, output.UserID)
}

func TestSetupUseCase_Execute_EmptyLogtoUserID(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)
	logtoClient := new(mockLogtoUpdater)

	uc := profile.NewSetupUseCase(userRepo, logtoClient)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrInvalidInput)
	require.Nil(t, output)
}

func TestSetupUseCase_Execute_AlreadyExists(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)
	logtoClient := new(mockLogtoUpdater)

	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(true, nil)

	uc := profile.NewSetupUseCase(userRepo, logtoClient)
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
	logtoClient := new(mockLogtoUpdater)

	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(false, errors.New("db unavailable"))

	uc := profile.NewSetupUseCase(userRepo, logtoClient)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "john",
	})

	require.Error(t, err)
	require.Nil(t, output)
}

func TestSetupUseCase_Execute_LogtoUpdateError(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)
	logtoClient := new(mockLogtoUpdater)

	logtoClient.On("UpdateUser", mock.Anything, "logto_user_1", "john", "", "", "").
		Return(errors.New("logto unavailable"))
	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(false, nil)

	uc := profile.NewSetupUseCase(userRepo, logtoClient)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "john",
	})

	require.Error(t, err)
	require.Nil(t, output)
}

func TestSetupUseCase_Execute_CreateDuplicate(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)
	logtoClient := new(mockLogtoUpdater)

	logtoClient.On("UpdateUser", mock.Anything, "logto_user_1", "john", "", "", "").
		Return(nil)
	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(false, nil)
	userRepo.EXPECT().
		Create(mock.Anything, "logto_user_1").
		Return(nil, errs.ErrDuplicate).
		Once()

	uc := profile.NewSetupUseCase(userRepo, logtoClient)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "john",
		Name:        "",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrDuplicate)
	require.Nil(t, output)
}

func TestSetupUseCase_Execute_CreateError(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)
	logtoClient := new(mockLogtoUpdater)

	logtoClient.On("UpdateUser", mock.Anything, "logto_user_1", "john", "", "", "").
		Return(nil)
	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_user_1").
		Return(false, nil)
	userRepo.EXPECT().
		Create(mock.Anything, "logto_user_1").
		Return(nil, errors.New("db unavailable")).
		Once()

	uc := profile.NewSetupUseCase(userRepo, logtoClient)
	output, err := uc.Execute(context.Background(), profile.SetupInput{
		LogtoUserID: "logto_user_1",
		Username:    "john",
	})

	require.Error(t, err)
	require.Nil(t, output)
}
