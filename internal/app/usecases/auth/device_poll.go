package auth

import (
	"context"
	"strings"

	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// PollDeviceTokenInput carries the device code to poll for tokens.
type PollDeviceTokenInput struct {
	DeviceCode string
}

// PollDeviceTokenOutput holds the OIDC tokens returned from a successful poll.
type PollDeviceTokenOutput struct {
	AccessToken  string
	IDToken      string
	RefreshToken string
	ExpiresIn    int
}

// PollDeviceTokenUseCase polls Logto for OIDC tokens after the user
// completes the device authorization flow.
type PollDeviceTokenUseCase struct {
	logtoDevice LogtoDeviceFlowClient
}

// NewPollDeviceTokenUseCase creates a new PollDeviceTokenUseCase.
func NewPollDeviceTokenUseCase(logtoDevice LogtoDeviceFlowClient) *PollDeviceTokenUseCase {
	return &PollDeviceTokenUseCase{
		logtoDevice: logtoDevice,
	}
}

// Execute polls Logto's token endpoint for the given device code.
//
// The caller should retry on ErrDeviceCodePending at the recommended interval.
// Other errors (ErrDeviceCodeExpired, ErrDeviceCodeAccessDenied) are terminal.
func (uc *PollDeviceTokenUseCase) Execute(
	ctx context.Context,
	input PollDeviceTokenInput,
) (*PollDeviceTokenOutput, error) {
	if strings.TrimSpace(input.DeviceCode) == "" {
		return nil, errs.ErrInvalidInput
	}

	resp, err := uc.logtoDevice.PollDeviceToken(ctx, input.DeviceCode)
	if err != nil {
		return nil, err // Pass through: pending, expired, denied, or unexpected
	}

	return &PollDeviceTokenOutput{
		AccessToken:  resp.AccessToken,
		IDToken:      resp.IDToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}
