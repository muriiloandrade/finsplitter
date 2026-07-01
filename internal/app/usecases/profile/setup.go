package profile

import (
	"context"
	"errors"
	"fmt"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// SetupInput carries the data needed to complete profile setup.
type SetupInput struct {
	LogtoUserID string
	Username    string
}

// SetupOutput holds the result of a successful profile setup.
type SetupOutput struct {
	UserID   string
	Username string
}

// SetupUseCase completes the profile setup for a user authenticated via Logto.
type SetupUseCase struct {
	userRepo ports.UserRepository
}

// NewSetupUseCase creates a new SetupUseCase.
func NewSetupUseCase(userRepo ports.UserRepository) *SetupUseCase {
	return &SetupUseCase{userRepo: userRepo}
}

// Execute creates a local Finsplitter user record linked to the Logto user.
// This is called after the user has authenticated via Logto for the first time.
func (uc *SetupUseCase) Execute(ctx context.Context, input SetupInput) (*SetupOutput, error) {
	if input.Username == "" {
		return nil, errs.ErrInvalidInput
	}

	exists, err := uc.userRepo.ExistsByLogtoUserID(ctx, input.LogtoUserID)
	if err != nil {
		return nil, fmt.Errorf("check user existence: %w", err)
	}
	if exists {
		return nil, errs.ErrDuplicate
	}

	user := &entity.User{
		LogtoUserID: input.LogtoUserID,
		Username:    input.Username,
	}

	if createErr := uc.userRepo.Create(ctx, user); createErr != nil {
		if errors.Is(createErr, errs.ErrDuplicate) {
			return nil, errs.ErrDuplicate
		}
		return nil, fmt.Errorf("create finsplitter user: %w", createErr)
	}

	return &SetupOutput{
		UserID:   user.ID.String(),
		Username: user.Username,
	}, nil
}
