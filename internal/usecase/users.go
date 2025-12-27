package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"desa-agent/internal/adapters"
	"desa-agent/internal/models"
)

type UsersUseCase struct {
	idp adapters.IdentityProvider
}

func NewUsersUseCase(idp adapters.IdentityProvider) *UsersUseCase {
	return &UsersUseCase{idp: idp}
}

func (uc *UsersUseCase) GetUser(ctx context.Context, userHash string, includePII bool) (*models.User, error) {
	user, err := uc.idp.GetUser(ctx, userHash)
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
	outCh := make(chan models.User)
	errCh := make(chan error, 1)

	go func() {
		defer close(outCh)
		defer close(errCh)

		usersCh, idpErrCh := uc.idp.ListUsers(ctx)

		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return

			case err, ok := <-idpErrCh:
				if ok && err != nil {
					errCh <- fmt.Errorf("idp error: %w", err)
					return
				}

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
