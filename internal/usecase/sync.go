package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"desa-agent/internal/models"
)

const syncInterval = 1 * time.Minute

func (u *UsersUseCase) StartSyncJob(ctx context.Context, logger *slog.Logger) {
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("sync job stopped")
			return
		case <-ticker.C:
			u.runSync(ctx, logger)
		}
	}
}

func (u *UsersUseCase) runSync(ctx context.Context, logger *slog.Logger) {
	logger.Info("starting user sync")
	if err := u.SyncUsers(ctx); err != nil {
		logger.Error("user sync failed", "error", err)
	} else {
		logger.Info("user sync completed successfully")
	}
}

func (u *UsersUseCase) SyncUsers(ctx context.Context) error {
	idpUsers, err := u.idp.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("idp.ListUsers: %w", err)
	}

	dbUsers, err := u.collectDBUsers(ctx)
	if err != nil {
		return fmt.Errorf("collectDBUsers: %w", err)
	}

	usersToUpsert := make([]models.User, 0)
	for _, idpUser := range idpUsers {
		dbUser, exists := dbUsers[idpUser.UserHash]
		if !exists || !reflect.DeepEqual(idpUser, dbUser) {
			usersToUpsert = append(usersToUpsert, idpUser)
		}
	}

	if len(usersToUpsert) == 0 {
		return nil
	}

	err = u.storage.UpsertUsers(ctx, usersToUpsert)
	if err != nil {
		return fmt.Errorf("storage.UpsertUsers: %w", err)
	}

	return nil
}

func (u *UsersUseCase) collectDBUsers(ctx context.Context) (map[string]models.User, error) {
	usersCh, errCh := u.storage.ListUsers(ctx)
	dbUsers := make(map[string]models.User)

	for user := range usersCh {
		dbUsers[user.UserHash] = user
	}

	if err := <-errCh; err != nil {
		return nil, err
	}

	return dbUsers, nil
}
