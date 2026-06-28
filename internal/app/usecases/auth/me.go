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
	Email       string
}

// MeOutput holds the current user's profile information.
type MeOutput struct {
	ID         string
	Username   string
	Email      string
	NeedsSetup bool
}

// MeUseCase returns the current user's profile, determining whether they still
// need to complete setup based on the existence of a local database record.
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
// the user needs setup, so we return only the email (available from JWT claims).
// If a record exists we return the full profile and NeedsSetup=false.
func (uc *MeUseCase) Execute(ctx context.Context, input MeInput) (*MeOutput, error) {
	user, err := uc.userRepo.GetByLogtoUserID(ctx, input.LogtoUserID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return &MeOutput{
				Email:      input.Email,
				NeedsSetup: true,
			}, nil
		}
		return nil, fmt.Errorf("get user by logto id: %w", err)
	}

	return &MeOutput{
		ID:         user.ID.String(),
		Username:   user.Username,
		Email:      input.Email,
		NeedsSetup: false,
	}, nil
}
