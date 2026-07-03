package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	profileUC "github.com/muriiloandrade/finsplitter/internal/app/usecases/profile"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockLogtoUpdater satisfies profileUC.LogtoUserUpdater.
type mockLogtoUpdater struct {
	mock.Mock
}

func (m *mockLogtoUpdater) UpdateUser(
	ctx context.Context, userID, username, name, phone, picture string,
) error {
	args := m.Called(ctx, userID, username, name, phone, picture)
	return args.Error(0)
}

// mockClaimsProvider returns the given claims from Claims().
type mockClaimsProvider struct {
	claims *entity.UserClaims
}

func (m *mockClaimsProvider) Claims(_ context.Context) *entity.UserClaims {
	return m.claims
}

func TestHandler_Setup_NoClaims(t *testing.T) {
	h := NewHandler(nil, &mockClaimsProvider{claims: nil})

	req := &SetupRequest{}
	req.Body.Username = "john"

	resp, err := h.Setup(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 401, statusErr.GetStatus())
	assert.Contains(t, statusErr.Error(), "unauthenticated")
}

func TestHandler_Setup_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV4())
	mockUserRepo := ports.NewMockUserRepository(t)
	mockUpdater := new(mockLogtoUpdater)

	mockUserRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_sub_1").
		Return(false, nil).
		Once()
	mockUpdater.On("UpdateUser", mock.Anything, "logto_sub_1", "john", "John", "+1234567890", "https://example.com/avatar.jpg").
		Return(nil)
	mockUserRepo.EXPECT().
		Create(mock.Anything, "logto_sub_1").
		Return(&entity.User{ID: userID, LogtoUserID: "logto_sub_1"}, nil).
		Once()

	setupUC := profileUC.NewSetupUseCase(mockUserRepo, mockUpdater)
	h := NewHandler(setupUC, &mockClaimsProvider{
		claims: &entity.UserClaims{Sub: "logto_sub_1"},
	})

	req := &SetupRequest{}
	req.Body.Username = "john"
	req.Body.Name = "John"
	req.Body.Phone = "+1234567890"
	req.Body.Picture = "https://example.com/avatar.jpg"

	resp, err := h.Setup(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, userID.String(), resp.Body.UserID)
	assert.Contains(t, resp.Body.Message, "Profile setup complete")
}

func TestHandler_Setup_Duplicate(t *testing.T) {
	mockUserRepo := ports.NewMockUserRepository(t)
	mockUpdater := new(mockLogtoUpdater)

	mockUserRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_sub_1").
		Return(true, nil).
		Once()

	setupUC := profileUC.NewSetupUseCase(mockUserRepo, mockUpdater)
	h := NewHandler(setupUC, &mockClaimsProvider{
		claims: &entity.UserClaims{Sub: "logto_sub_1"},
	})

	req := &SetupRequest{}
	req.Body.Username = "john"

	resp, err := h.Setup(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 409, statusErr.GetStatus())
	assert.Contains(t, statusErr.Error(), "profile already set up")
}

func TestHandler_Setup_InvalidInput(t *testing.T) {
	setupUC := profileUC.NewSetupUseCase(nil, nil)
	h := NewHandler(setupUC, &mockClaimsProvider{
		claims: &entity.UserClaims{Sub: ""},
	})

	req := &SetupRequest{}
	req.Body.Username = "john"

	// The use case returns ErrInvalidInput when LogtoUserID is empty
	// (even though the handler checks claims != nil, it still passes the empty Sub)
	// Wait - the handler checks claims == nil, not claims.Sub == "".
	// So we need a different approach.

	// Actually the handler calls:
	//   claims := h.claimsPr.Claims(ctx)
	//   if claims == nil { return 401 }
	//   output, err := h.setupUC.Execute(ctx, profile.SetupInput{
	//       LogtoUserID: claims.Sub,
	//       ...
	//   })
	//
	// If claims.Sub is empty, the use case returns ErrInvalidInput.

	resp, err := h.Setup(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 422, statusErr.GetStatus())
	assert.Contains(t, statusErr.Error(), "invalid input")
}

func TestHandler_Setup_GenericError(t *testing.T) {
	mockUserRepo := ports.NewMockUserRepository(t)
	mockUpdater := new(mockLogtoUpdater)

	mockUserRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "logto_sub_1").
		Return(false, errors.New("db connection error")).
		Once()

	setupUC := profileUC.NewSetupUseCase(mockUserRepo, mockUpdater)
	h := NewHandler(setupUC, &mockClaimsProvider{
		claims: &entity.UserClaims{Sub: "logto_sub_1"},
	})

	req := &SetupRequest{}
	req.Body.Username = "john"

	resp, err := h.Setup(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)

	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, 500, statusErr.GetStatus())
	assert.Contains(t, statusErr.Error(), "setup failed")
}
