package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	authUC "github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestHandler_Register_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV4())
	mockUserRepo := ports.NewMockUserRepository(t)
	mockCreator := authUC.NewMockLogtoUserCreator(t)

	mockCreator.EXPECT().CreateUser(mock.Anything, "john", "", "John", "john@example.com").
		Return(&logto.CreateUserResponse{ID: "logto_user_1"}, nil)
	mockUserRepo.EXPECT().
		Create(mock.Anything, "logto_user_1").
		Return(&entity.User{ID: userID, LogtoUserID: "logto_user_1"}, nil).
		Once()

	registerUC := authUC.NewRegisterUseCase(mockUserRepo, mockCreator)
	h := NewHandler(registerUC, nil, nil, nil)

	req := &RegisterRequest{}
	req.Body.Name = "John"
	req.Body.Email = "john@example.com"
	req.Body.Username = "john"

	resp, err := h.Register(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, userID.String(), resp.Body.UserID)
	assert.NotEmpty(t, resp.Body.Message, "registration confirmation message")

	mockCreator.AssertExpectations(t)
}

func TestHandler_Register_UsernameTaken(t *testing.T) {
	mockUserRepo := ports.NewMockUserRepository(t)
	mockCreator := authUC.NewMockLogtoUserCreator(t)

	mockCreator.EXPECT().CreateUser(mock.Anything, "john", "", "John", "john@example.com").
		Return(nil, logto.ErrUserExists)

	registerUC := authUC.NewRegisterUseCase(mockUserRepo, mockCreator)
	h := NewHandler(registerUC, nil, nil, nil)

	req := &RegisterRequest{}
	req.Body.Name = "John"
	req.Body.Email = "john@example.com"
	req.Body.Username = "john"

	resp, err := h.Register(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 409, statusErr.GetStatus())
	assert.Contains(t, statusErr.Error(), "username already taken")

	mockCreator.AssertExpectations(t)
}

func TestHandler_Register_UserAlreadyExists(t *testing.T) {
	mockUserRepo := ports.NewMockUserRepository(t)
	mockCreator := authUC.NewMockLogtoUserCreator(t)

	mockCreator.EXPECT().CreateUser(mock.Anything, "john", "", "John", "john@example.com").
		Return(&logto.CreateUserResponse{ID: "logto_user_1"}, nil)
	mockUserRepo.EXPECT().
		Create(mock.Anything, "logto_user_1").
		Return(&entity.User{}, errs.ErrDuplicate).
		Once()

	registerUC := authUC.NewRegisterUseCase(mockUserRepo, mockCreator)
	h := NewHandler(registerUC, nil, nil, nil)

	req := &RegisterRequest{}
	req.Body.Name = "John"
	req.Body.Email = "john@example.com"
	req.Body.Username = "john"

	resp, err := h.Register(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 409, statusErr.GetStatus())
	assert.Contains(t, statusErr.Error(), "user already registered")

	mockCreator.AssertExpectations(t)
}

func TestHandler_Register_GenericError(t *testing.T) {
	mockUserRepo := ports.NewMockUserRepository(t)
	mockCreator := authUC.NewMockLogtoUserCreator(t)

	mockCreator.EXPECT().CreateUser(mock.Anything, "john", "", "John", "john@example.com").
		Return(nil, errors.New("unexpected logto error"))

	registerUC := authUC.NewRegisterUseCase(mockUserRepo, mockCreator)
	h := NewHandler(registerUC, nil, nil, nil)

	req := &RegisterRequest{}
	req.Body.Name = "John"
	req.Body.Email = "john@example.com"
	req.Body.Username = "john"

	resp, err := h.Register(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 500, statusErr.GetStatus())
	assert.Contains(t, statusErr.Error(), "registration failed")

	mockCreator.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Me
// ---------------------------------------------------------------------------

func TestHandler_Me_NoClaims(t *testing.T) {
	h := NewHandler(nil, nil, nil, nil)

	resp, err := h.Me(context.Background(), &struct{}{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Body.ID)
	assert.False(t, resp.Body.NeedsSetup)
}

func TestHandler_Me_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV4())
	mockUserRepo := ports.NewMockUserRepository(t)

	mockUserRepo.EXPECT().
		GetByLogtoUserID(mock.Anything, "logto_sub_1").
		Return(&entity.User{ID: userID}, nil).
		Once()

	meUC := authUC.NewMeUseCase(mockUserRepo)
	h := NewHandler(nil, meUC, nil, nil)

	ctx := WithUserClaims(context.Background(), &entity.UserClaims{
		Sub:      "logto_sub_1",
		Username: "john",
		Email:    "john@example.com",
		Name:     "John",
		Phone:    "+1234567890",
		Picture:  "https://example.com/avatar.jpg",
	})

	resp, err := h.Me(ctx, &struct{}{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, userID.String(), resp.Body.ID)
	assert.Equal(t, "john", resp.Body.Username)
	assert.Equal(t, "john@example.com", resp.Body.Email)
	assert.Equal(t, "John", resp.Body.Name)
	assert.Equal(t, "+1234567890", resp.Body.Phone)
	assert.Equal(t, "https://example.com/avatar.jpg", resp.Body.Picture)
	assert.False(t, resp.Body.NeedsSetup)

	mockUserRepo.AssertExpectations(t)
}

func TestHandler_Me_NeedsSetup(t *testing.T) {
	mockUserRepo := ports.NewMockUserRepository(t)

	mockUserRepo.EXPECT().
		GetByLogtoUserID(mock.Anything, "logto_sub_new").
		Return(nil, errs.ErrNotFound).
		Once()

	meUC := authUC.NewMeUseCase(mockUserRepo)
	h := NewHandler(nil, meUC, nil, nil)

	ctx := WithUserClaims(context.Background(), &entity.UserClaims{
		Sub: "logto_sub_new",
	})

	resp, err := h.Me(ctx, &struct{}{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Body.ID)
	assert.True(t, resp.Body.NeedsSetup)

	mockUserRepo.AssertExpectations(t)
}

func TestHandler_Me_UseCaseError(t *testing.T) {
	mockUserRepo := ports.NewMockUserRepository(t)

	mockUserRepo.EXPECT().
		GetByLogtoUserID(mock.Anything, "logto_sub_1").
		Return(nil, errors.New("db connection error")).
		Once()

	meUC := authUC.NewMeUseCase(mockUserRepo)
	h := NewHandler(nil, meUC, nil, nil)

	ctx := WithUserClaims(context.Background(), &entity.UserClaims{
		Sub: "logto_sub_1",
	})

	resp, err := h.Me(ctx, &struct{}{})

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 500, statusErr.GetStatus())
	assert.Contains(t, statusErr.Error(), "failed to get user info")

	mockUserRepo.AssertExpectations(t)
}
