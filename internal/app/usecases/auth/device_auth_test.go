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
