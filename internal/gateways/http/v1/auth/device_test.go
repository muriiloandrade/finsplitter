package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// DeviceAuth (POST /auth/device)
// ---------------------------------------------------------------------------

func TestHandler_DeviceAuth_Success(t *testing.T) {
	mockLogtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	mockLogtoDevice.EXPECT().RequestDeviceCode(mock.Anything).
		Return(&logto.DeviceCodeResponse{
			DeviceCode:              "dc_123",
			UserCode:                "ABCD-EFGH",
			VerificationURI:         "http://localhost:3001/device",
			VerificationURIComplete: "http://localhost:3001/device?user_code=ABCD-EFGH",
			ExpiresIn:               1800,
			Interval:                5,
		}, nil)

	deviceAuthUC := auth.NewRequestDeviceAuthUseCase(mockLogtoDevice)
	h := &Handler{deviceAuthUC: deviceAuthUC}

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
	mockLogtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	// Empty email is caught by the use case before calling the Logto client.
	deviceAuthUC := auth.NewRequestDeviceAuthUseCase(mockLogtoDevice)
	h := &Handler{deviceAuthUC: deviceAuthUC}

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
	mockLogtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	mockLogtoDevice.EXPECT().RequestDeviceCode(mock.Anything).
		Return(nil, errors.New("logto unavailable"))

	deviceAuthUC := auth.NewRequestDeviceAuthUseCase(mockLogtoDevice)
	h := &Handler{deviceAuthUC: deviceAuthUC}

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
	mockLogtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	mockLogtoDevice.EXPECT().PollDeviceToken(mock.Anything, "dc_123").
		Return(&logto.DeviceTokenResponse{
			AccessToken:  "access_abc",
			IDToken:      "id_def",
			RefreshToken: "refresh_ghi",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}, nil)

	devicePollUC := auth.NewPollDeviceTokenUseCase(mockLogtoDevice)
	h := &Handler{devicePollUC: devicePollUC}

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
	mockLogtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	mockLogtoDevice.EXPECT().PollDeviceToken(mock.Anything, "dc_pending").
		Return(nil, logto.ErrDeviceCodePending)

	devicePollUC := auth.NewPollDeviceTokenUseCase(mockLogtoDevice)
	h := &Handler{devicePollUC: devicePollUC}

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
	mockLogtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	mockLogtoDevice.EXPECT().PollDeviceToken(mock.Anything, "dc_expired").
		Return(nil, logto.ErrDeviceCodeExpired)

	devicePollUC := auth.NewPollDeviceTokenUseCase(mockLogtoDevice)
	h := &Handler{devicePollUC: devicePollUC}

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
	mockLogtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	mockLogtoDevice.EXPECT().PollDeviceToken(mock.Anything, "dc_denied").
		Return(nil, logto.ErrDeviceCodeAccessDenied)

	devicePollUC := auth.NewPollDeviceTokenUseCase(mockLogtoDevice)
	h := &Handler{devicePollUC: devicePollUC}

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
	mockLogtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	mockLogtoDevice.EXPECT().PollDeviceToken(mock.Anything, "dc_fail").
		Return(nil, errors.New("unexpected error"))

	devicePollUC := auth.NewPollDeviceTokenUseCase(mockLogtoDevice)
	h := &Handler{devicePollUC: devicePollUC}

	req := &PollDeviceTokenRequest{}
	req.Body.DeviceCode = "dc_fail"

	resp, err := h.DevicePoll(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 500, statusErr.GetStatus())
}

func TestHandler_DevicePoll_InvalidInput(t *testing.T) {
	mockLogtoDevice := auth.NewMockLogtoDeviceFlowClient(t)

	// Empty DeviceCode is caught by the use case before calling the mock,
	// so no expectations on mockLogtoDevice are needed.
	devicePollUC := auth.NewPollDeviceTokenUseCase(mockLogtoDevice)
	h := &Handler{devicePollUC: devicePollUC}

	req := &PollDeviceTokenRequest{}
	// Empty DeviceCode triggers errs.ErrInvalidInput in the use case.
	req.Body.DeviceCode = ""

	resp, err := h.DevicePoll(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, http.StatusUnprocessableEntity, statusErr.GetStatus())
}
