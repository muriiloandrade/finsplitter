package usecases

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type UpdateUserUC interface {
	UpdateUser(ctx context.Context, id uuid.UUID, user entity.User) (entity.User, error)
}

type UpdateUserUseCase struct {
	repo ports.UpdateUserRepository
}

func NewUpdateUserUC(repo ports.UpdateUserRepository) *UpdateUserUseCase {
	return &UpdateUserUseCase{repo: repo}
}

func (uc *UpdateUserUseCase) UpdateUser(ctx context.Context, id uuid.UUID, user entity.User) (entity.User, error) {
	return uc.repo.UpdateUser(ctx, id, user)
}
