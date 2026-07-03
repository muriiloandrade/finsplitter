package auth

import (
	"context"
	"errors"
	"fmt"
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

// mockLogtoCreator satisfies authUC.LogtoUserCreator.
type mockLogtoCreator struct {
	mock.Mock
}

func (m *mockLogtoCreator) CreateUser(
	ctx context.Context, username, password, name, email string,
) (*logto.CreateUserResponse, error) {
	args := m.Called(ctx, username, password, name, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*logto.CreateUserResponse), args.Error(1)
}

// mockLogtoAuthenticator satisfies authUC.LogtoUserAuthenticator.
type mockLogtoAuthenticator struct {
	mock.Mock
}

func (m *mockLogtoAuthenticator) AuthenticateUser(
	ctx context.Context, email, password string,
) (*logto.TokenResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*logto.TokenResponse), args.Error(1)
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestHandler_Register_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV4())
	mockUserRepo := ports.NewMockUserRepository(t)
	mockCreator := new(mockLogtoCreator)

	mockCreator.On("CreateUser", mock.Anything, "john", "secret123", "John", "john@example.com").
		Return(&logto.CreateUserResponse{ID: "logto_user_1"}, nil)
	mockUserRepo.EXPECT().
		Create(mock.Anything, "logto_user_1").
		Return(&entity.User{ID: userID, LogtoUserID: "logto_user_1"}, nil).
		Once()

	registerUC := authUC.NewRegisterUseCase(mockUserRepo, mockCreator)
	h := NewHandler(registerUC, nil, nil)

	req := &RegisterRequest{}
	req.Body.Name = "John"
	req.Body.Email = "john@example.com"
	req.Body.Username = "john"
	req.Body.Password = "secret123"

	resp, err := h.Register(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, userID.String(), resp.Body.UserID)
	assert.Equal(t, "/auth/sign-in", resp.Body.RedirectURL)

	mockCreator.AssertExpectations(t)
}

func TestHandler_Register_UsernameTaken(t *testing.T) {
	mockUserRepo := ports.NewMockUserRepository(t)
	mockCreator := new(mockLogtoCreator)

	mockCreator.On("CreateUser", mock.Anything, "john", "secret123", "John", "john@example.com").
		Return(nil, logto.ErrUserExists)

	registerUC := authUC.NewRegisterUseCase(mockUserRepo, mockCreator)
	h := NewHandler(registerUC, nil, nil)

	req := &RegisterRequest{}
	req.Body.Name = "John"
	req.Body.Email = "john@example.com"
	req.Body.Username = "john"
	req.Body.Password = "secret123"

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
	mockCreator := new(mockLogtoCreator)

	mockCreator.On("CreateUser", mock.Anything, "john", "secret123", "John", "john@example.com").
		Return(&logto.CreateUserResponse{ID: "logto_user_1"}, nil)
	mockUserRepo.EXPECT().
		Create(mock.Anything, "logto_user_1").
		Return(nil, errs.ErrDuplicate).
		Once()

	registerUC := authUC.NewRegisterUseCase(mockUserRepo, mockCreator)
	h := NewHandler(registerUC, nil, nil)

	req := &RegisterRequest{}
	req.Body.Name = "John"
	req.Body.Email = "john@example.com"
	req.Body.Username = "john"
	req.Body.Password = "secret123"

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
	mockCreator := new(mockLogtoCreator)

	mockCreator.On("CreateUser", mock.Anything, "john", "secret123", "John", "john@example.com").
		Return(nil, errors.New("unexpected logto error"))

	registerUC := authUC.NewRegisterUseCase(mockUserRepo, mockCreator)
	h := NewHandler(registerUC, nil, nil)

	req := &RegisterRequest{}
	req.Body.Name = "John"
	req.Body.Email = "john@example.com"
	req.Body.Username = "john"
	req.Body.Password = "secret123"

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
	h := NewHandler(nil, nil, nil)

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
	h := NewHandler(nil, meUC, nil)

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
	h := NewHandler(nil, meUC, nil)

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
	h := NewHandler(nil, meUC, nil)

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

// ---------------------------------------------------------------------------
// SignIn
// ---------------------------------------------------------------------------

func TestHandler_SignIn_Success(t *testing.T) {
	mockAuthenticator := new(mockLogtoAuthenticator)

	mockAuthenticator.On("AuthenticateUser", mock.Anything, "john@example.com", "secret123").
		Return(&logto.TokenResponse{
			AccessToken:  "access_abc",
			IDToken:      "id_def",
			RefreshToken: "refresh_ghi",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}, nil)

	signInUC := authUC.NewSignInUseCase(mockAuthenticator)
	h := NewHandler(nil, nil, signInUC)

	req := &SignInRequest{}
	req.Body.Email = "john@example.com"
	req.Body.Password = "secret123"

	resp, err := h.SignIn(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "access_abc", resp.Body.AccessToken)
	assert.Equal(t, "id_def", resp.Body.IDToken)
	assert.Equal(t, "refresh_ghi", resp.Body.RefreshToken)
	assert.Equal(t, 3600, resp.Body.ExpiresIn)

	mockAuthenticator.AssertExpectations(t)
}

func TestHandler_SignIn_InvalidCredentials(t *testing.T) {
	mockAuthenticator := new(mockLogtoAuthenticator)

	mockAuthenticator.On("AuthenticateUser", mock.Anything, "wrong@example.com", "badpass").
		Return(nil, fmt.Errorf("authenticate user: %w", errs.ErrInvalidCredentials))

	signInUC := authUC.NewSignInUseCase(mockAuthenticator)
	h := NewHandler(nil, nil, signInUC)

	req := &SignInRequest{}
	req.Body.Email = "wrong@example.com"
	req.Body.Password = "badpass"

	resp, err := h.SignIn(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 401, statusErr.GetStatus())
	assert.Contains(t, statusErr.Error(), "invalid email or password")

	mockAuthenticator.AssertExpectations(t)
}

func TestHandler_SignIn_GenericError(t *testing.T) {
	mockAuthenticator := new(mockLogtoAuthenticator)

	mockAuthenticator.On("AuthenticateUser", mock.Anything, "john@example.com", "secret123").
		Return(nil, errors.New("logto network error"))

	signInUC := authUC.NewSignInUseCase(mockAuthenticator)
	h := NewHandler(nil, nil, signInUC)

	req := &SignInRequest{}
	req.Body.Email = "john@example.com"
	req.Body.Password = "secret123"

	resp, err := h.SignIn(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 500, statusErr.GetStatus())
	assert.Contains(t, statusErr.Error(), "sign-in failed")

	mockAuthenticator.AssertExpectations(t)
}
