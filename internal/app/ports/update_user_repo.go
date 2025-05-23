package ports

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

type UpdateUserRepository interface {
	UpdateUser(ctx context.Context, id uuid.UUID, user entity.User) (entity.User, error)
}
