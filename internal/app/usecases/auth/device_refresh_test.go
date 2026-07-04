package auth_test

import (
	"context"
	"errors"
	"testing"

	auth "github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// RefreshDeviceTokenUseCase
// ---------------------------------------------------------------------------

func TestRefreshDeviceTokenUseCase_Execute_Success(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().RefreshDeviceToken(mock.Anything, "refresh_token_abc").
		Return(&logto.DeviceTokenRefreshResponse{
			AccessToken:  "new_access_token",
			IDToken:      "new_id_token",
			RefreshToken: "new_refresh_token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}, nil)

	uc := auth.NewRefreshDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.RefreshDeviceTokenInput{
		RefreshToken: "refresh_token_abc",
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "new_access_token", output.AccessToken)
	assert.Equal(t, "new_id_token", output.IDToken)
	assert.Equal(t, "new_refresh_token", output.RefreshToken)
	assert.Equal(t, 3600, output.ExpiresIn)
}

func TestRefreshDeviceTokenUseCase_Execute_EmptyToken(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)
	uc := auth.NewRefreshDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.RefreshDeviceTokenInput{
		RefreshToken: "",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrInvalidInput)
	require.Nil(t, output)
}

func TestRefreshDeviceTokenUseCase_Execute_Expired(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().RefreshDeviceToken(mock.Anything, "expired_refresh_token").
		Return(nil, logto.ErrDeviceCodeExpired)

	uc := auth.NewRefreshDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.RefreshDeviceTokenInput{
		RefreshToken: "expired_refresh_token",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, logto.ErrDeviceCodeExpired)
	require.Nil(t, output)
}

func TestRefreshDeviceTokenUseCase_Execute_LogtoError(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().RefreshDeviceToken(mock.Anything, "refresh_token_abc").
		Return(nil, errors.New("logto unavailable"))

	uc := auth.NewRefreshDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.RefreshDeviceTokenInput{
		RefreshToken: "refresh_token_abc",
	})

	require.Error(t, err)
	require.Nil(t, output)
}
