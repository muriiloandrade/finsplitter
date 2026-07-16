package auth

import (
	"context"
	"strings"

	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// RevokeDeviceTokenInput carries the refresh token to revoke.
type RevokeDeviceTokenInput struct {
	RefreshToken string
}

// RevokeDeviceTokenUseCase revokes a device flow refresh token via Logto.
type RevokeDeviceTokenUseCase struct {
	logtoDevice LogtoDeviceFlowClient
}

// NewRevokeDeviceTokenUseCase creates a new RevokeDeviceTokenUseCase.
func NewRevokeDeviceTokenUseCase(logtoDevice LogtoDeviceFlowClient) *RevokeDeviceTokenUseCase {
	return &RevokeDeviceTokenUseCase{
		logtoDevice: logtoDevice,
	}
}

// Execute revokes the given refresh token.
// Returns ErrInvalidInput if refreshToken is empty.
func (uc *RevokeDeviceTokenUseCase) Execute(
	ctx context.Context,
	input RevokeDeviceTokenInput,
) error {
	if strings.TrimSpace(input.RefreshToken) == "" {
		return errs.ErrInvalidInput
	}

	return uc.logtoDevice.RevokeDeviceToken(ctx, input.RefreshToken)
}
