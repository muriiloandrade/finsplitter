package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// RequestDeviceAuthInput carries the data needed to start a device auth flow.
type RequestDeviceAuthInput struct {
	Email string
}

// RequestDeviceAuthOutput holds the device auth response from Logto.
type RequestDeviceAuthOutput struct {
	DeviceCode              string
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	ExpiresIn               int
	Interval                int
}

// RequestDeviceAuthUseCase initiates the device authorization flow via Logto.
type RequestDeviceAuthUseCase struct {
	logtoDevice LogtoDeviceFlowClient
}

// NewRequestDeviceAuthUseCase creates a new RequestDeviceAuthUseCase.
func NewRequestDeviceAuthUseCase(logtoDevice LogtoDeviceFlowClient) *RequestDeviceAuthUseCase {
	return &RequestDeviceAuthUseCase{
		logtoDevice: logtoDevice,
	}
}

// Execute requests a device code from Logto for the given email.
func (uc *RequestDeviceAuthUseCase) Execute(
	ctx context.Context,
	input RequestDeviceAuthInput,
) (*RequestDeviceAuthOutput, error) {
	if strings.TrimSpace(input.Email) == "" {
		return nil, errs.ErrInvalidInput
	}

	resp, err := uc.logtoDevice.RequestDeviceCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("request device code: %w", err)
	}

	return &RequestDeviceAuthOutput{
		DeviceCode:              resp.DeviceCode,
		UserCode:                resp.UserCode,
		VerificationURI:         resp.VerificationURI,
		VerificationURIComplete: resp.VerificationURIComplete,
		ExpiresIn:               resp.ExpiresIn,
		Interval:                resp.Interval,
	}, nil
}
