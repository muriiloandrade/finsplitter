package auth_test

import (
	"context"
	"errors"
	"testing"

	auth "github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockLogtoAuthenticator is a manual mock for auth.LogtoUserAuthenticator.
type mockLogtoAuthenticator struct {
	mock.Mock
}

func (m *mockLogtoAuthenticator) AuthenticateUser(
	ctx context.Context, email, password string,
) (*logto.TokenResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*logto.TokenResponse), args.Error(1)
}

func TestSignInUseCase_Execute_Success(t *testing.T) {
	logtoAuth := new(mockLogtoAuthenticator)

	logtoAuth.On("AuthenticateUser", mock.Anything, "john@example.com", "secret123").
		Return(&logto.TokenResponse{
			AccessToken:  "access_token_123",
			IDToken:      "id_token_456",
			RefreshToken: "refresh_token_789",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}, nil)

	uc := auth.NewSignInUseCase(logtoAuth)
	output, err := uc.Execute(context.Background(), auth.SignInInput{
		Email:    "john@example.com",
		Password: "secret123",
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "access_token_123", output.AccessToken)
	assert.Equal(t, "id_token_456", output.IDToken)
	assert.Equal(t, "refresh_token_789", output.RefreshToken)
	assert.Equal(t, 3600, output.ExpiresIn)
}

func TestSignInUseCase_Execute_InvalidCredentials(t *testing.T) {
	logtoAuth := new(mockLogtoAuthenticator)

	logtoAuth.On("AuthenticateUser", mock.Anything, "wrong@example.com", "badpass").
		Return(nil, logto.ErrInvalidCredentials)

	uc := auth.NewSignInUseCase(logtoAuth)
	output, err := uc.Execute(context.Background(), auth.SignInInput{
		Email:    "wrong@example.com",
		Password: "badpass",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrInvalidCredentials)
	require.Nil(t, output)
}

func TestSignInUseCase_Execute_LogtoError(t *testing.T) {
	logtoAuth := new(mockLogtoAuthenticator)

	logtoAuth.On("AuthenticateUser", mock.Anything, "john@example.com", "secret123").
		Return(nil, errors.New("logto unavailable"))

	uc := auth.NewSignInUseCase(logtoAuth)
	output, err := uc.Execute(context.Background(), auth.SignInInput{
		Email:    "john@example.com",
		Password: "secret123",
	})

	require.Error(t, err)
	require.Nil(t, output)
}

func TestSignInUseCase_Execute_EmptyEmail(t *testing.T) {
	uc := auth.NewSignInUseCase(new(mockLogtoAuthenticator))
	output, err := uc.Execute(context.Background(), auth.SignInInput{
		Email:    "",
		Password: "secret123",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrInvalidCredentials)
	require.Nil(t, output)
}

func TestSignInUseCase_Execute_EmptyPassword(t *testing.T) {
	uc := auth.NewSignInUseCase(new(mockLogtoAuthenticator))
	output, err := uc.Execute(context.Background(), auth.SignInInput{
		Email:    "john@example.com",
		Password: "",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrInvalidCredentials)
	require.Nil(t, output)
}
