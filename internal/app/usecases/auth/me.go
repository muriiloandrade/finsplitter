package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// MeInput carries the data needed to look up the current user.
type MeInput struct {
	LogtoUserID string
}

// MeOutput holds the current user's Finsplitter account status.
// Profile data (name, email, username) is available from JWT claims directly.
type MeOutput struct {
	ID         string
	NeedsSetup bool
}

// MeUseCase returns the current user's Finsplitter account status, determining
// whether they still need to complete setup based on the existence of a local
// database record.
type MeUseCase struct {
	userRepo ports.UserRepository
}

// NewMeUseCase creates a new MeUseCase.
func NewMeUseCase(userRepo ports.UserRepository) *MeUseCase {
	return &MeUseCase{
		userRepo: userRepo,
	}
}

// Execute looks up the user by their Logto ID. If no local record exists yet
// the user needs setup. If a record exists we return the Finsplitter user ID
// and NeedsSetup=false.
func (uc *MeUseCase) Execute(ctx context.Context, input MeInput) (*MeOutput, error) {
	user, err := uc.userRepo.GetByLogtoUserID(ctx, input.LogtoUserID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return &MeOutput{
				NeedsSetup: true,
			}, nil
		}
		return nil, fmt.Errorf("get user by logto id: %w", err)
	}

	return &MeOutput{
		ID:         user.ID.String(),
		NeedsSetup: false,
	}, nil
}
