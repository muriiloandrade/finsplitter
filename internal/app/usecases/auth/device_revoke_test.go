package auth

import (
	"context"
	"testing"

	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRevokeDeviceTokenUseCase_Execute_Success(t *testing.T) {
	mockClient := NewMockLogtoDeviceFlowClient(t)
	uc := NewRevokeDeviceTokenUseCase(mockClient)

	mockClient.EXPECT().RevokeDeviceToken(mock.Anything, "valid-refresh-token").Return(nil).Once()

	err := uc.Execute(context.Background(), RevokeDeviceTokenInput{RefreshToken: "valid-refresh-token"})

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRevokeDeviceTokenUseCase_Execute_EmptyToken(t *testing.T) {
	uc := NewRevokeDeviceTokenUseCase(NewMockLogtoDeviceFlowClient(t))

	err := uc.Execute(context.Background(), RevokeDeviceTokenInput{RefreshToken: ""})

	assert.ErrorIs(t, err, errs.ErrInvalidInput)
}

func TestRevokeDeviceTokenUseCase_Execute_WhitespaceToken(t *testing.T) {
	uc := NewRevokeDeviceTokenUseCase(NewMockLogtoDeviceFlowClient(t))

	err := uc.Execute(context.Background(), RevokeDeviceTokenInput{RefreshToken: "   "})

	assert.ErrorIs(t, err, errs.ErrInvalidInput)
}

func TestRevokeDeviceTokenUseCase_Execute_LogtoError(t *testing.T) {
	mockClient := NewMockLogtoDeviceFlowClient(t)
	uc := NewRevokeDeviceTokenUseCase(mockClient)

	mockClient.EXPECT().RevokeDeviceToken(mock.Anything, "bad-token").Return(logto.ErrAppClientNotConfigured).Once()

	err := uc.Execute(context.Background(), RevokeDeviceTokenInput{RefreshToken: "bad-token"})

	assert.ErrorIs(t, err, logto.ErrAppClientNotConfigured)
}
