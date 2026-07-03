package ports

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

// UserRepository defines the contract for user data access.
type UserRepository interface {
	// Create inserts a new user linked to the given Logto user ID.
	// Returns ErrDuplicate if logto_user_id already exists.
	Create(ctx context.Context, logtoUserID string) (*entity.User, error)
	// GetByID retrieves a user by their Finsplitter ID.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	// GetByLogtoUserID retrieves a user by their Logto user ID.
	GetByLogtoUserID(ctx context.Context, logtoUserID string) (*entity.User, error)
	// ExistsByLogtoUserID checks whether a user with the given Logto user ID exists.
	ExistsByLogtoUserID(ctx context.Context, logtoUserID string) (bool, error)
}
