package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

// RegisterInput carries the data needed to register a new user.
type RegisterInput struct {
	Username string
	Password string
}

// RegisterOutput holds the result of a successful registration.
type RegisterOutput struct {
	UserID      string
	LogtoUserID string
}

// RegisterUseCase orchestrates user registration in Logto and Finsplitter.
type RegisterUseCase struct {
	userRepo ports.UserRepository
	logtoM2M *logto.Client
}

// NewRegisterUseCase creates a new RegisterUseCase.
func NewRegisterUseCase(userRepo ports.UserRepository, logtoM2M *logto.Client) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo: userRepo,
		logtoM2M: logtoM2M,
	}
}

// Execute creates a user in Logto via the Management API, then persists a local
// user record. If the Logto creation succeeds but the local creation fails, the
// Logto user is NOT rolled back (Logto is the source of truth for identity).
func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	logtoUser, err := uc.logtoM2M.CreateUser(ctx, input.Username, input.Password)
	if err != nil {
		if errors.Is(err, logto.ErrUserExists) {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("create logto user: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate user id: %w", err)
	}

	user := &entity.User{
		ID:          id,
		LogtoUserID: logtoUser.ID,
		Username:    input.Username,
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
	}, nil
}
