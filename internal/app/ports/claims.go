package ports

import (
	"context"

	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
)

// ClaimsProvider provides access to the authenticated user's claims
// extracted from the JWT and stored in the request context.
type ClaimsProvider interface {
	Claims(ctx context.Context) *entity.UserClaims
}
