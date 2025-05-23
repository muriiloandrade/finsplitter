package ports

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type DeleteUserRepository interface {
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
