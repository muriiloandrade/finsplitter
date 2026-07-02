package profile

import (
	"context"
	"errors"
	"fmt"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// SetupOutput holds the result of a successful profile setup.
type SetupOutput struct {
	UserID string
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
// The optional username from the request is not stored locally — profile data
// lives in Logto and is read from JWT claims.
func (uc *SetupUseCase) Execute(ctx context.Context, logtoUserID string) (*SetupOutput, error) {
	if logtoUserID == "" {
		return nil, errs.ErrInvalidInput
	}

	exists, err := uc.userRepo.ExistsByLogtoUserID(ctx, logtoUserID)
	if err != nil {
		return nil, fmt.Errorf("check user existence: %w", err)
	}
	if exists {
		return nil, errs.ErrDuplicate
	}

	user, createErr := uc.userRepo.Create(ctx, logtoUserID)
	if createErr != nil {
		if errors.Is(createErr, errs.ErrDuplicate) {
			return nil, errs.ErrDuplicate
		}
		return nil, fmt.Errorf("create finsplitter user: %w", createErr)
	}

	return &SetupOutput{
		UserID: user.ID.String(),
	}, nil
}
