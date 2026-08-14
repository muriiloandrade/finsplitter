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
// PollDeviceTokenUseCase
// ---------------------------------------------------------------------------

func TestPollDeviceTokenUseCase_Execute_Success(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().PollDeviceToken(mock.Anything, "device_code_abc").
		Return(&logto.DeviceTokenResponse{
			AccessToken:  "access_token_123",
			IDToken:      "id_token_456",
			RefreshToken: "refresh_token_789",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}, nil)

	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "device_code_abc",
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "access_token_123", output.AccessToken)
	assert.Equal(t, "id_token_456", output.IDToken)
	assert.Equal(t, "refresh_token_789", output.RefreshToken)
	assert.Equal(t, 3600, output.ExpiresIn)
}

func TestPollDeviceTokenUseCase_Execute_EmptyDeviceCode(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)
	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)

	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "",
	})

	require.ErrorIs(t, err, errs.ErrInvalidInput)
	assert.Nil(t, output)
}

func TestPollDeviceTokenUseCase_Execute_WhitespaceDeviceCode(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)
	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)

	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "   ",
	})

	require.ErrorIs(t, err, errs.ErrInvalidInput)
	assert.Nil(t, output)
}

func TestPollDeviceTokenUseCase_Execute_Pending(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().PollDeviceToken(mock.Anything, "device_code_abc").
		Return(nil, logto.ErrDeviceCodePending)

	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "device_code_abc",
	})

	require.ErrorIs(t, err, logto.ErrDeviceCodePending)
	assert.Nil(t, output)
}

func TestPollDeviceTokenUseCase_Execute_Expired(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().PollDeviceToken(mock.Anything, "device_code_abc").
		Return(nil, logto.ErrDeviceCodeExpired)

	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "device_code_abc",
	})

	require.ErrorIs(t, err, logto.ErrDeviceCodeExpired)
	assert.Nil(t, output)
}

func TestPollDeviceTokenUseCase_Execute_AccessDenied(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	logtoDevice.EXPECT().PollDeviceToken(mock.Anything, "device_code_abc").
		Return(nil, logto.ErrDeviceCodeAccessDenied)

	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "device_code_abc",
	})

	require.ErrorIs(t, err, logto.ErrDeviceCodeAccessDenied)
	assert.Nil(t, output)
}

func TestPollDeviceTokenUseCase_Execute_UnexpectedError(t *testing.T) {
	logtoDevice := auth.NewMockLogtoDeviceFlowClient(t)
	unexpected := errors.New("logto unavailable")

	logtoDevice.EXPECT().PollDeviceToken(mock.Anything, "device_code_abc").
		Return(nil, unexpected)

	uc := auth.NewPollDeviceTokenUseCase(logtoDevice)
	output, err := uc.Execute(context.Background(), auth.PollDeviceTokenInput{
		DeviceCode: "device_code_abc",
	})

	require.ErrorIs(t, err, unexpected)
	assert.Nil(t, output)
}
