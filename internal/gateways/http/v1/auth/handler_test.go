package auth_test

import (
	"context"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	authUC "github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	authHandler "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockSignInUseCase is a manual mock for the sign-in use case.
type mockSignInUseCase struct {
	mock.Mock
}

func (m *mockSignInUseCase) Execute(ctx context.Context, input authUC.SignInInput) (*authUC.SignInOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authUC.SignInOutput), args.Error(1)
}

func TestSignInHandler_SignIn(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		email     string
		password  string
		ucSetup   func(uc *mockSignInUseCase)
		wantErr   bool
		errStatus int
		wantBody  struct {
			accessToken string
			idToken     string
			expiresIn   int
		}
	}{
		{
			name:     "success",
			email:    "john@example.com",
			password: "secret123",
			ucSetup: func(uc *mockSignInUseCase) {
				uc.On("Execute", mock.Anything, authUC.SignInInput{
					Email:    "john@example.com",
					Password: "secret123",
				}).Return(&authUC.SignInOutput{
					AccessToken:  "access_token_123",
					IDToken:      "id_token_456",
					RefreshToken: "refresh_token_789",
					ExpiresIn:    3600,
				}, nil)
			},
			wantErr: false,
			wantBody: struct {
				accessToken string
				idToken     string
				expiresIn   int
			}{
				accessToken: "access_token_123",
				idToken:     "id_token_456",
				expiresIn:   3600,
			},
		},
		{
			name:     "returns 401 on invalid credentials",
			email:    "wrong@example.com",
			password: "badpass",
			ucSetup: func(uc *mockSignInUseCase) {
				uc.On("Execute", mock.Anything, authUC.SignInInput{
					Email:    "wrong@example.com",
					Password: "badpass",
				}).Return(nil, authUC.ErrSignInInvalidCredentials)
			},
			wantErr:   true,
			errStatus: 401,
		},
		{
			name:     "returns 500 on generic error",
			email:    "john@example.com",
			password: "secret123",
			ucSetup: func(uc *mockSignInUseCase) {
				uc.On("Execute", mock.Anything, authUC.SignInInput{
					Email:    "john@example.com",
					Password: "secret123",
				}).Return(nil, assert.AnError)
			},
			wantErr:   true,
			errStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mockSignInUseCase)
			tt.ucSetup(mockUC)

			h := authHandler.NewHandler(nil, nil, mockUC)

			req := &authHandler.SignInRequest{}
			req.Body.Email = tt.email
			req.Body.Password = tt.password

			resp, err := h.SignIn(ctx, req)

			if tt.wantErr {
				require.Error(t, err)
				var statusErr huma.StatusError
				require.ErrorAs(t, err, &statusErr)
				assert.Equal(t, tt.errStatus, statusErr.GetStatus())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.wantBody.accessToken, resp.Body.AccessToken)
				assert.Equal(t, tt.wantBody.idToken, resp.Body.IDToken)
				assert.Equal(t, tt.wantBody.expiresIn, resp.Body.ExpiresIn)
				if tt.wantBody.idToken == "" {
					assert.Empty(t, resp.Body.RefreshToken)
				}
			}

			mockUC.AssertExpectations(t)
		})
	}
}
