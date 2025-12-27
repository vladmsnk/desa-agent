package adapters

import (
	"context"

	"desa-agent/internal/models"
)

type IdentityProvider interface {
	GetUser(ctx context.Context, userID string) (*models.User, error)
	ListUsers(ctx context.Context) (<-chan models.User, <-chan error)
	Close() error
}
