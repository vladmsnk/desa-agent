package ad

import (
	"context"
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
	return nil, nil
}

func (a *Adapter) ListUsers(ctx context.Context) ([]models.User, error) {
	return nil, nil
}

func (a *Adapter) Close() error {
	return nil
}
