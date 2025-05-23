package ports

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type ListUserRepository interface {
	ListUsers(ctx context.Context) ([]entity.User, error)
}
