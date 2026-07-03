package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

// Compile-time check that *logto.Client satisfies LogtoUserAuthenticator.
var _ LogtoUserAuthenticator = (*logto.Client)(nil)

// SignInInput carries the credentials needed to authenticate a user.
type SignInInput struct {
	Email    string
	Password string
}

// SignInOutput holds the tokens returned from a successful sign-in.
type SignInOutput struct {
	AccessToken  string
	IDToken      string
	RefreshToken string
	ExpiresIn    int
}

// SignInUseCase authenticates a user against Logto's OIDC provider using
// the Resource Owner Password Credentials (ROPC) grant.
//
// On success it returns the tokens that the client can use to authenticate
// subsequent requests to Finsplitter.
type SignInUseCase struct {
	logtoAuth LogtoUserAuthenticator
}

// NewSignInUseCase creates a new SignInUseCase.
func NewSignInUseCase(logtoAuth LogtoUserAuthenticator) *SignInUseCase {
	return &SignInUseCase{
		logtoAuth: logtoAuth,
	}
}

// Execute authenticates the user with Logto using the provided credentials
// and returns OIDC tokens.
func (uc *SignInUseCase) Execute(ctx context.Context, input SignInInput) (*SignInOutput, error) {
	if strings.TrimSpace(input.Email) == "" || input.Password == "" {
		return nil, errs.ErrInvalidCredentials
	}

	tokenResp, err := uc.logtoAuth.AuthenticateUser(ctx, input.Email, input.Password)
	if err != nil {
		if errors.Is(err, logto.ErrInvalidCredentials) {
			return nil, errs.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("authenticate user: %w", err)
	}

	return &SignInOutput{
		AccessToken:  tokenResp.AccessToken,
		IDToken:      tokenResp.IDToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}
