package usecases

import (
	"context"
	"log/slog"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	slogctx "github.com/veqryn/slog-context"
)

type ListUserUC interface {
	ListUsers(ctx context.Context) ([]entity.User, error)
}

type ListUsersUseCase struct {
	repo ports.ListUserRepository
}

func NewListUserUC(repo ports.ListUserRepository) *ListUsersUseCase {
	return &ListUsersUseCase{repo: repo}
}

func (uc *ListUsersUseCase) ListUsers(ctx context.Context) ([]entity.User, error) {
	logger := slogctx.FromCtx(ctx)
	users, err := uc.repo.ListUsers(ctx)
	if err != nil {
		logger.Error("Failed to list users", slog.Any("error", err))
		return nil, err
	}
	return users, nil
}
