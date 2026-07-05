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
// RequestDeviceAuthUseCase
// ---------------------------------------------------------------------------

func TestRequestDeviceAuthUseCase_Execute_Success(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().RequestDeviceCode(mock.Anything).
		Return(&logto.DeviceCodeResponse{
			DeviceCode:              "device_code_123",
			UserCode:                "ABCD-EFGH",
			VerificationURI:         "http://localhost:3001/device",
			VerificationURIComplete: "http://localhost:3001/device?user_code=ABCD-EFGH",
			ExpiresIn:               1800,
			Interval:                5,
		}, nil)

	uc := auth.NewRequestDeviceAuthUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.RequestDeviceAuthInput{
		Email: "user@example.com",
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "device_code_123", output.DeviceCode)
	assert.Equal(t, "ABCD-EFGH", output.UserCode)
	assert.Equal(t, "http://localhost:3001/device", output.VerificationURI)
	assert.Equal(t, "http://localhost:3001/device?user_code=ABCD-EFGH", output.VerificationURIComplete)
	assert.Equal(t, 1800, output.ExpiresIn)
	assert.Equal(t, 5, output.Interval)
}

func TestRequestDeviceAuthUseCase_Execute_EmptyEmail(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)
	uc := auth.NewRequestDeviceAuthUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.RequestDeviceAuthInput{
		Email: "",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrInvalidInput)
	require.Nil(t, output)
}

func TestRequestDeviceAuthUseCase_Execute_LogtoError(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().RequestDeviceCode(mock.Anything).
		Return(nil, errors.New("logto unavailable"))

	uc := auth.NewRequestDeviceAuthUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.RequestDeviceAuthInput{
		Email: "user@example.com",
	})

	require.Error(t, err)
	require.Nil(t, output)
}

// ---------------------------------------------------------------------------
// PollDeviceTokenUseCase
// ---------------------------------------------------------------------------

func TestPollDeviceTokenUseCase_Execute_Success(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().PollDeviceToken(mock.Anything, "device_code_123").
		Return(&logto.DeviceTokenResponse{
			AccessToken:  "access_token_abc",
			IDToken:      "id_token_def",
			RefreshToken: "refresh_token_ghi",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}, nil)

	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "device_code_123",
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "access_token_abc", output.AccessToken)
	assert.Equal(t, "id_token_def", output.IDToken)
	assert.Equal(t, "refresh_token_ghi", output.RefreshToken)
	assert.Equal(t, 3600, output.ExpiresIn)
}

func TestPollDeviceTokenUseCase_Execute_Pending(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().PollDeviceToken(mock.Anything, "device_code_123").
		Return(nil, logto.ErrDeviceCodePending)

	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "device_code_123",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, logto.ErrDeviceCodePending)
	require.Nil(t, output)
}

func TestPollDeviceTokenUseCase_Execute_Expired(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().PollDeviceToken(mock.Anything, "device_code_expired").
		Return(nil, logto.ErrDeviceCodeExpired)

	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "device_code_expired",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, logto.ErrDeviceCodeExpired)
	require.Nil(t, output)
}

func TestPollDeviceTokenUseCase_Execute_AccessDenied(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().PollDeviceToken(mock.Anything, "device_code_denied").
		Return(nil, logto.ErrDeviceCodeAccessDenied)

	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "device_code_denied",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, logto.ErrDeviceCodeAccessDenied)
	require.Nil(t, output)
}

func TestPollDeviceTokenUseCase_Execute_EmptyDeviceCode(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)
	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrInvalidInput)
	require.Nil(t, output)
}
