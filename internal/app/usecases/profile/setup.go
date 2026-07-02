package profile

import (
	"context"
	"errors"
	"fmt"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// LogtoUserUpdater updates user profile fields in Logto via the Management API.
// Satisfied by *logto.Client in production.
type LogtoUserUpdater interface {
	UpdateUser(ctx context.Context, userID, username, phone, picture string) error
}

// SetupInput carries the data needed to complete profile setup.
type SetupInput struct {
	LogtoUserID string
	Username    string
	Phone       string
	Picture     string
}

// SetupOutput holds the result of a successful profile setup.
type SetupOutput struct {
	UserID string
}

// SetupUseCase completes the profile setup for a user authenticated via Logto.
// It creates a local Finsplitter user record and updates the Logto user's
// profile with the provided fields (username, phone, avatar).
type SetupUseCase struct {
	userRepo    ports.UserRepository
	logtoClient LogtoUserUpdater
}

// NewSetupUseCase creates a new SetupUseCase.
func NewSetupUseCase(userRepo ports.UserRepository, logtoClient LogtoUserUpdater) *SetupUseCase {
	return &SetupUseCase{
		userRepo:    userRepo,
		logtoClient: logtoClient,
	}
}

// Execute creates a local Finsplitter user record linked to the Logto user and
// updates the Logto profile with the provided username, phone, and picture.
func (uc *SetupUseCase) Execute(ctx context.Context, input SetupInput) (*SetupOutput, error) {
	if input.LogtoUserID == "" {
		return nil, errs.ErrInvalidInput
	}

	exists, err := uc.userRepo.ExistsByLogtoUserID(ctx, input.LogtoUserID)
	if err != nil {
		return nil, fmt.Errorf("check user existence: %w", err)
	}
	if exists {
		return nil, errs.ErrDuplicate
	}

	// Update Logto profile first (best-effort — if this fails the user can retry setup).
	if updateErr := uc.logtoClient.UpdateUser(
		ctx,
		input.LogtoUserID,
		input.Username,
		input.Phone,
		input.Picture,
	); updateErr != nil {
		return nil, fmt.Errorf("update logto profile: %w", updateErr)
	}

	user, createErr := uc.userRepo.Create(ctx, input.LogtoUserID)
	if createErr != nil {
		// Logto user was updated but local record failed — Logto is source of truth,
		// the user can retry setup to create the local record.
		if errors.Is(createErr, errs.ErrDuplicate) {
			return nil, errs.ErrDuplicate
		}
		return nil, fmt.Errorf("create finsplitter user: %w", createErr)
	}

	return &SetupOutput{
		UserID: user.ID.String(),
	}, nil
}
