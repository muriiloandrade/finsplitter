package auth

import (
	"context"
	"strings"

	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// RefreshDeviceTokenInput carries the refresh token to exchange for new tokens.
type RefreshDeviceTokenInput struct {
	RefreshToken string
}

// RefreshDeviceTokenOutput holds the new tokens from a successful refresh.
type RefreshDeviceTokenOutput struct {
	AccessToken  string
	IDToken      string
	RefreshToken string
	ExpiresIn    int
}

// RefreshDeviceTokenUseCase refreshes OIDC tokens using a refresh token.
type RefreshDeviceTokenUseCase struct {
	logtoDevice LogtoDeviceFlowClient
}

// NewRefreshDeviceTokenUseCase creates a new RefreshDeviceTokenUseCase.
func NewRefreshDeviceTokenUseCase(logtoDevice LogtoDeviceFlowClient) *RefreshDeviceTokenUseCase {
	return &RefreshDeviceTokenUseCase{
		logtoDevice: logtoDevice,
	}
}

// Execute exchanges the refresh token for new access and refresh tokens.
//
// Logto rotates refresh tokens, so callers MUST store the returned
// refresh_token for subsequent refreshes.
func (uc *RefreshDeviceTokenUseCase) Execute(
	ctx context.Context,
	input RefreshDeviceTokenInput,
) (*RefreshDeviceTokenOutput, error) {
	if strings.TrimSpace(input.RefreshToken) == "" {
		return nil, errs.ErrInvalidInput
	}

	resp, err := uc.logtoDevice.RefreshDeviceToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, err // Pass through: expired, or unexpected
	}

	return &RefreshDeviceTokenOutput{
		AccessToken:  resp.AccessToken,
		IDToken:      resp.IDToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}
