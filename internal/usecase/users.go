package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"desa-agent/internal/models"
)

type Storage interface {
	GetUser(ctx context.Context, userHash string) (*models.User, error)
	ListUsers(ctx context.Context) (<-chan models.User, <-chan error)
	UpsertUsers(ctx context.Context, users []models.User) error
}

type IdentityProvider interface {
	GetUser(ctx context.Context, userID string) (*models.User, error)
	ListUsers(ctx context.Context) ([]models.User, error)
}

type UsersUseCase struct {
	storage Storage
	idp     IdentityProvider
}

func NewUsersUseCase(storage Storage, idp IdentityProvider) *UsersUseCase {
	return &UsersUseCase{storage: storage, idp: idp}
}

func (uc *UsersUseCase) GetUser(ctx context.Context, userHash string, includePII bool) (*models.User, error) {
	user, err := uc.storage.GetUser(ctx, userHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, nil
	}

	if !includePII {
		user.PII = nil
	}

	return user, nil
}

func (uc *UsersUseCase) ListUsers(ctx context.Context, includePII bool) (<-chan models.User, <-chan error) {
	usersCh, storageErrCh := uc.storage.ListUsers(ctx)

	outCh := make(chan models.User)
	errCh := make(chan error, 1)

	go func() {
		defer close(outCh)
		defer close(errCh)

		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return

			case err, ok := <-storageErrCh:
				if ok && err != nil {
					errCh <- fmt.Errorf("storage error: %w", err)
				}
				return

			case user, ok := <-usersCh:
				if !ok {
					return
				}

				if !includePII {
					user.PII = nil
				}

				select {
				case outCh <- user:
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			}
		}
	}()

	return outCh, errCh
}

func HashUserID(sourceID string) string {
	hash := sha256.Sum256([]byte(sourceID))
	return hex.EncodeToString(hash[:])
}
