package ldap

import (
	"context"
	"fmt"

	"desa-agent/internal/config"
	"desa-agent/internal/models"
)

type Adapter struct {
	cfg config.IDPConfig
}

func New(cfg config.IDPConfig) (*Adapter, error) {
	return &Adapter{cfg: cfg}, nil
}

func (a *Adapter) GetUser(ctx context.Context, userID string) (*models.User, error) {
	return nil, fmt.Errorf("LDAP adapter: GetUser not implemented")
}

func (a *Adapter) ListUsers(ctx context.Context) (<-chan models.User, <-chan error) {
	usersCh := make(chan models.User)
	errCh := make(chan error, 1)

	go func() {
		defer close(usersCh)
		defer close(errCh)
		errCh <- fmt.Errorf("LDAP adapter: ListUsers not implemented")
	}()

	return usersCh, errCh
}

func (a *Adapter) Close() error {
	return nil
}
