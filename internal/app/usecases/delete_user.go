package usecases

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
)

type DeleteUserUC interface {
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type DeleteUserUseCase struct {
	repo ports.DeleteUserRepository
}

func NewDeleteUserUC(repo ports.DeleteUserRepository) *DeleteUserUseCase {
	return &DeleteUserUseCase{repo: repo}
}

func (uc *DeleteUserUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return uc.repo.DeleteUser(ctx, id)
}
