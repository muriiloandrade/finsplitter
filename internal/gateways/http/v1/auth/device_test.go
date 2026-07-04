package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// DeviceAuth (POST /auth/device)
// ---------------------------------------------------------------------------

func TestHandler_DeviceAuth_Success(t *testing.T) {
	mockAuthUC := newMockdeviceAuthUseCase(t)

	mockAuthUC.EXPECT().Execute(mock.Anything, auth.RequestDeviceAuthInput{
		Email: "user@example.com",
	}).Return(&auth.RequestDeviceAuthOutput{
		DeviceCode:              "dc_123",
		UserCode:                "ABCD-EFGH",
		VerificationURI:         "http://localhost:3001/device",
		VerificationURIComplete: "http://localhost:3001/device?user_code=ABCD-EFGH",
		ExpiresIn:               1800,
		Interval:                5,
	}, nil)

	h := &Handler{deviceAuthUC: mockAuthUC}

	req := &RequestDeviceAuthRequest{}
	req.Body.Email = "user@example.com"

	resp, err := h.DeviceAuth(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "dc_123", resp.Body.DeviceCode)
	assert.Equal(t, "ABCD-EFGH", resp.Body.UserCode)
	assert.Equal(t, "http://localhost:3001/device", resp.Body.VerificationURI)
	assert.Equal(t, "http://localhost:3001/device?user_code=ABCD-EFGH", resp.Body.VerificationURIComplete)
	assert.Equal(t, 1800, resp.Body.ExpiresIn)
	assert.Equal(t, 5, resp.Body.Interval)
}

func TestHandler_DeviceAuth_InvalidInput(t *testing.T) {
	mockAuthUC := newMockdeviceAuthUseCase(t)

	mockAuthUC.EXPECT().Execute(mock.Anything, auth.RequestDeviceAuthInput{
		Email: "",
	}).Return(nil, errs.ErrInvalidInput)

	h := &Handler{deviceAuthUC: mockAuthUC}

	req := &RequestDeviceAuthRequest{}
	req.Body.Email = ""

	resp, err := h.DeviceAuth(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 422, statusErr.GetStatus())
}

func TestHandler_DeviceAuth_GenericError(t *testing.T) {
	mockAuthUC := newMockdeviceAuthUseCase(t)

	mockAuthUC.EXPECT().Execute(mock.Anything, auth.RequestDeviceAuthInput{
		Email: "user@example.com",
	}).Return(nil, errors.New("logto unavailable"))

	h := &Handler{deviceAuthUC: mockAuthUC}

	req := &RequestDeviceAuthRequest{}
	req.Body.Email = "user@example.com"

	resp, err := h.DeviceAuth(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 500, statusErr.GetStatus())
}

// ---------------------------------------------------------------------------
// DevicePoll (POST /auth/device/poll)
// ---------------------------------------------------------------------------

func TestHandler_DevicePoll_Success(t *testing.T) {
	mockPollUC := newMockdevicePollUseCase(t)

	mockPollUC.EXPECT().Execute(mock.Anything, auth.PollDeviceTokenInput{
		DeviceCode: "dc_123",
	}).Return(&auth.PollDeviceTokenOutput{
		AccessToken:  "access_abc",
		IDToken:      "id_def",
		RefreshToken: "refresh_ghi",
		ExpiresIn:    3600,
	}, nil)

	h := &Handler{devicePollUC: mockPollUC}

	req := &PollDeviceTokenRequest{}
	req.Body.DeviceCode = "dc_123"

	resp, err := h.DevicePoll(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "access_abc", resp.Body.AccessToken)
	assert.Equal(t, "id_def", resp.Body.IDToken)
	assert.Equal(t, "refresh_ghi", resp.Body.RefreshToken)
	assert.Equal(t, 3600, resp.Body.ExpiresIn)
}

func TestHandler_DevicePoll_Pending(t *testing.T) {
	mockPollUC := newMockdevicePollUseCase(t)

	mockPollUC.EXPECT().Execute(mock.Anything, auth.PollDeviceTokenInput{
		DeviceCode: "dc_pending",
	}).Return(nil, logto.ErrDeviceCodePending)

	h := &Handler{devicePollUC: mockPollUC}

	req := &PollDeviceTokenRequest{}
	req.Body.DeviceCode = "dc_pending"

	resp, err := h.DevicePoll(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, http.StatusUnauthorized, statusErr.GetStatus())
}

func TestHandler_DevicePoll_Expired(t *testing.T) {
	mockPollUC := newMockdevicePollUseCase(t)

	mockPollUC.EXPECT().Execute(mock.Anything, auth.PollDeviceTokenInput{
		DeviceCode: "dc_expired",
	}).Return(nil, logto.ErrDeviceCodeExpired)

	h := &Handler{devicePollUC: mockPollUC}

	req := &PollDeviceTokenRequest{}
	req.Body.DeviceCode = "dc_expired"

	resp, err := h.DevicePoll(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 400, statusErr.GetStatus())
}

func TestHandler_DevicePoll_AccessDenied(t *testing.T) {
	mockPollUC := newMockdevicePollUseCase(t)

	mockPollUC.EXPECT().Execute(mock.Anything, auth.PollDeviceTokenInput{
		DeviceCode: "dc_denied",
	}).Return(nil, logto.ErrDeviceCodeAccessDenied)

	h := &Handler{devicePollUC: mockPollUC}

	req := &PollDeviceTokenRequest{}
	req.Body.DeviceCode = "dc_denied"

	resp, err := h.DevicePoll(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 403, statusErr.GetStatus())
}

func TestHandler_DevicePoll_GenericError(t *testing.T) {
	mockPollUC := newMockdevicePollUseCase(t)

	mockPollUC.EXPECT().Execute(mock.Anything, auth.PollDeviceTokenInput{
		DeviceCode: "dc_fail",
	}).Return(nil, errors.New("unexpected error"))

	h := &Handler{devicePollUC: mockPollUC}

	req := &PollDeviceTokenRequest{}
	req.Body.DeviceCode = "dc_fail"

	resp, err := h.DevicePoll(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 500, statusErr.GetStatus())
}
