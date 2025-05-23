package ports

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type CreateUserRepository interface {
	CreateUser(ctx context.Context, user entity.User) (entity.User, error)
}
