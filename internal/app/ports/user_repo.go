package ports

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

// UserRepository defines the contract for user data access.
type UserRepository interface {
	// Create inserts a new user. Returns ErrDuplicate if logto_user_id already exists.
	Create(ctx context.Context, user *entity.User) error
	// GetByID retrieves a user by their Finsplitter ID.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	// GetByLogtoUserID retrieves a user by their Logto user ID.
	GetByLogtoUserID(ctx context.Context, logtoUserID string) (*entity.User, error)
	// UpdateUsername updates the username for an existing user.
	UpdateUsername(ctx context.Context, id uuid.UUID, username string) error
	// ExistsByLogtoUserID checks whether a user with the given Logto user ID exists.
	ExistsByLogtoUserID(ctx context.Context, logtoUserID string) (bool, error)
	// FindUsernamesByPrefix returns usernames that start with the given prefix.
	// The prefix should include the trailing '%' wildcard for the LIKE query.
	FindUsernamesByPrefix(ctx context.Context, prefix string) ([]string, error)
}
