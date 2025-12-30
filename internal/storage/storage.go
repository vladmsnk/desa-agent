package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"desa-agent/internal/models"
	"github.com/dgraph-io/badger/v4"
)

const userKeyPrefix = "user:"

type Storage struct {
	db *badger.DB
}

type Config struct {
	Path     string
	InMemory bool
}

func New(cfg Config) (*Storage, error) {
	opts := badger.DefaultOptions(cfg.Path)

	if cfg.InMemory {
		opts = opts.WithInMemory(true)
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) GetUser(ctx context.Context, userHash string) (*models.User, error) {
	var user models.User

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(userKeyPrefix + userHash))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &user)
		})
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *Storage) ListUsers(ctx context.Context) (<-chan models.User, <-chan error) {
	usersCh := make(chan models.User)
	errCh := make(chan error, 1)

	go func() {
		defer close(usersCh)
		defer close(errCh)

		err := s.db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = []byte(userKeyPrefix)

			it := txn.NewIterator(opts)
			defer it.Close()

			for it.Rewind(); it.Valid(); it.Next() {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				var user models.User
				err := it.Item().Value(func(val []byte) error {
					return json.Unmarshal(val, &user)
				})
				if err != nil {
					return err
				}

				select {
				case usersCh <- user:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})

		if err != nil {
			errCh <- err
		}
	}()

	return usersCh, errCh
}

func (s *Storage) UpsertUsers(ctx context.Context, users []models.User) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		for _, user := range users {
			data, err := json.Marshal(user)
			if err != nil {
				return fmt.Errorf("failed to marshal user %s: %w", user.UserHash, err)
			}

			if err := txn.Set([]byte(userKeyPrefix+user.UserHash), data); err != nil {
				return fmt.Errorf("failed to set user %s: %w", user.UserHash, err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to upsert users: %w", err)
	}

	return nil
}

func (s *Storage) RemoveUser(ctx context.Context, userHash string) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(userKeyPrefix + userHash))
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to remove user: %w", err)
	}

	return nil
}
