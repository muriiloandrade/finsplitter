package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	auth "github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockLogtoCreator is a manual mock for auth.LogtoUserCreator.
type mockLogtoCreator struct {
	mock.Mock
}

func (m *mockLogtoCreator) CreateUser(
	ctx context.Context, username, password string,
) (*logto.CreateUserResponse, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*logto.CreateUserResponse), args.Error(1)
}

func TestRegisterUseCase_Execute_Success(t *testing.T) {
	logtoM2M := new(mockLogtoCreator)
	userRepo := ports.NewMockUserRepository(t)

	logtoM2M.On("CreateUser", mock.Anything, "john", "secret123").
		Return(&logto.CreateUserResponse{ID: "logto_user_1", Username: "john"}, nil)
	userRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(nil).
		Once()

	uc := auth.NewRegisterUseCase(userRepo, logtoM2M)
	output, err := uc.Execute(context.Background(), auth.RegisterInput{
		Username: "john",
		Password: "secret123",
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "logto_user_1", output.LogtoUserID)
	assert.NotEmpty(t, output.UserID)
}

func TestRegisterUseCase_Execute_LogtoUserExists(t *testing.T) {
	logtoM2M := new(mockLogtoCreator)
	userRepo := ports.NewMockUserRepository(t)

	logtoM2M.On("CreateUser", mock.Anything, "john", "secret123").
		Return(nil, logto.ErrUserExists)

	uc := auth.NewRegisterUseCase(userRepo, logtoM2M)
	output, err := uc.Execute(context.Background(), auth.RegisterInput{
		Username: "john",
		Password: "secret123",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, auth.ErrUsernameTaken)
	require.Nil(t, output)
}

func TestRegisterUseCase_Execute_LogtoError(t *testing.T) {
	logtoM2M := new(mockLogtoCreator)
	userRepo := ports.NewMockUserRepository(t)

	logtoM2M.On("CreateUser", mock.Anything, "john", "secret123").
		Return(nil, errors.New("logto unavailable"))

	uc := auth.NewRegisterUseCase(userRepo, logtoM2M)
	output, err := uc.Execute(context.Background(), auth.RegisterInput{
		Username: "john",
		Password: "secret123",
	})

	require.Error(t, err)
	require.Nil(t, output)
}

func TestRegisterUseCase_Execute_LocalDuplicate(t *testing.T) {
	logtoM2M := new(mockLogtoCreator)
	userRepo := ports.NewMockUserRepository(t)

	logtoM2M.On("CreateUser", mock.Anything, "john", "secret123").
		Return(&logto.CreateUserResponse{ID: "logto_user_1", Username: "john"}, nil)
	userRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(errs.ErrDuplicate).
		Once()

	uc := auth.NewRegisterUseCase(userRepo, logtoM2M)
	output, err := uc.Execute(context.Background(), auth.RegisterInput{
		Username: "john",
		Password: "secret123",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, auth.ErrUserAlreadyExists)
	require.Nil(t, output)
}

func TestRegisterUseCase_Execute_LocalCreateError(t *testing.T) {
	logtoM2M := new(mockLogtoCreator)
	userRepo := ports.NewMockUserRepository(t)

	logtoM2M.On("CreateUser", mock.Anything, "john", "secret123").
		Return(&logto.CreateUserResponse{ID: "logto_user_1", Username: "john"}, nil)
	userRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(errors.New("db unavailable")).
		Once()

	uc := auth.NewRegisterUseCase(userRepo, logtoM2M)
	output, err := uc.Execute(context.Background(), auth.RegisterInput{
		Username: "john",
		Password: "secret123",
	})

	require.Error(t, err)
	require.Nil(t, output)
}
