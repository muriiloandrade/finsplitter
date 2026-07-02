package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
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
	Password string
}

// RegisterOutput holds the result of a successful registration.
type RegisterOutput struct {
	UserID      string
	LogtoUserID string
	Username    string
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
// user record. If the Logto creation succeeds but the local creation fails, the
// Logto user is NOT rolled back (Logto is the source of truth for identity).
func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	username := input.Username
	if username == "" {
		username = slugify(input.Name)
		// Find the next available username by checking existing prefixes.
		existing, err := uc.userRepo.FindUsernamesByPrefix(ctx, username+"%")
		if err != nil {
			return nil, fmt.Errorf("check username prefix: %w", err)
		}
		username = nextAvailableUsername(username, existing)
	}

	logtoUser, err := uc.logtoM2M.CreateUser(ctx, username, input.Password)
	if err != nil {
		if errors.Is(err, logto.ErrUserExists) {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("create logto user: %w", err)
	}

	user := &entity.User{
		LogtoUserID: logtoUser.ID,
		Username:    username,
		Name:        input.Name,
		Email:       input.Email,
	}

	if createErr := uc.userRepo.Create(ctx, user); createErr != nil {
		if errors.Is(createErr, errs.ErrDuplicate) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create finsplitter user: %w", createErr)
	}

	return &RegisterOutput{
		UserID:      user.ID.String(),
		LogtoUserID: user.LogtoUserID,
		Username:    username,
	}, nil
}

// nextAvailableUsername finds the next unused username from a prefix + existing list.
// If the base slug is unused, returns it as-is.
// Otherwise appends the next incrementing suffix: slug-1, slug-2, etc.
func nextAvailableUsername(slug string, existing []string) string {
	// Check if slug itself is available
	for _, u := range existing {
		if u == slug {
			goto taken
		}
	}
	return slug

taken:
	maxSuffix := 0
	for _, u := range existing {
		var suffix int
		n, _ := fmt.Sscanf(u, slug+"_%d", &suffix)
		if n == 1 && suffix > maxSuffix {
			maxSuffix = suffix
		}
	}
	return fmt.Sprintf("%s_%d", slug, maxSuffix+1)
}

// slugify converts a display name into a Logto-compatible username.
// Logto accepts lowercase alphanumeric and underscores.
// Separator is not used for names that differ only by non-alphanumeric chars
// (e.g. "John-Doe" and "John Doe" both produce "john_doe").
// Examples: "John Doe" → "john_doe", "  Hello   World! " → "hello_world".
func slugify(s string) string {
	var b strings.Builder
	prevSep := true // trim leading separators
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSep = false
		} else if !prevSep {
			b.WriteRune('_')
			prevSep = true
		}
	}
	return strings.TrimRight(b.String(), "_")
}
