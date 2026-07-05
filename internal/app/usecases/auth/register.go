package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

// Compile-time check that *logto.Client satisfies LogtoUserCreator.
var _ LogtoUserCreator = (*logto.Client)(nil)

// RegisterInput carries the data needed to register a new user.
type RegisterInput struct {
	Name     string
	Email    string
	Username string // optional — auto-generated from Name when empty
}

// RegisterOutput holds the result of a successful registration.
type RegisterOutput struct {
	UserID      string
	LogtoUserID string
}

// RegisterUseCase orchestrates user registration in Logto and Finsplitter.
type RegisterUseCase struct {
	userRepo ports.UserRepository
	logtoM2M LogtoUserCreator
}

// NewRegisterUseCase creates a new RegisterUseCase.
func NewRegisterUseCase(userRepo ports.UserRepository, logtoM2M LogtoUserCreator) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo: userRepo,
		logtoM2M: logtoM2M,
	}
}

// Execute creates a user in Logto via the Management API, then persists a local
// user record (ID-only link). If the Logto creation succeeds but the local
// creation fails, the Logto user is NOT rolled back (Logto is the source of
// truth for identity).
func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	username := input.Username
	if username == "" {
		username = slugify(input.Name)
	}

	// Passwordless registration: empty string tells Logto to create a passwordless user.
	logtoUser, err := uc.logtoM2M.CreateUser(ctx, username, "", input.Name, input.Email)
	if err != nil {
		if errors.Is(err, logto.ErrUserExists) {
			return nil, errs.ErrUsernameTaken
		}
		if errors.Is(err, logto.ErrEmailAlreadyInUse) {
			return nil, errs.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create logto user: %w", err)
	}

	user, createErr := uc.userRepo.Create(ctx, logtoUser.ID)
	if createErr != nil {
		if errors.Is(createErr, errs.ErrDuplicate) {
			return nil, errs.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create finsplitter user: %w", createErr)
	}

	return &RegisterOutput{
		UserID:      user.ID.String(),
		LogtoUserID: user.LogtoUserID,
	}, nil
}

// slugify converts a display name into a Logto-compatible username.
// Logto accepts only lowercase alphanumeric characters.
// Examples: "John Doe" → "johndoe", "  Hello   World! " → "helloworld".
// Returns empty string when no valid characters remain, which causes the
// username field to be omitted entirely from the Management API request.
func slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
