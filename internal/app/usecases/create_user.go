package usecases

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type CreateUserUC interface {
	CreateUser(ctx context.Context, user entity.User) (entity.User, error)
}

type CreateUserUseCase struct {
	repo ports.CreateUserRepository
}

func NewCreateUserUC(repo ports.CreateUserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{repo: repo}
}

func (uc *CreateUserUseCase) CreateUser(ctx context.Context, user entity.User) (entity.User, error) {
	return uc.repo.CreateUser(ctx, user)
}
