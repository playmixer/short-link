package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/database"
	"github.com/playmixer/short-link/internal/adapters/storage/file"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
)

type Config struct {
	Memory   *memory.Config
	File     *file.Config
	Database *database.Config
}

type Store interface {
	Set(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
	SetBatch(ctx context.Context, batch []models.ShortLink) error
}

func NewStore(cfg *Config) (Store, error) {
	if cfg.Database != nil && cfg.Database.DSN != "" {
		store, err := database.New(cfg.Database)
		if err != nil {
			return nil, fmt.Errorf("failed initialize database storage: %w", err)
		}
		return store, nil
	}

	if cfg.File != nil && cfg.File.StoragePath != "" {
		store, err := file.New(cfg.File)
		if err != nil {
			return nil, fmt.Errorf("failed initialize file storage: %w", err)
		}
		return store, nil
	}

	if cfg.Memory != nil {
		store, err := memory.New(cfg.Memory)
		if err != nil {
			return nil, fmt.Errorf("failed initialize memory storage: %w", err)
		}
		return store, nil
	}

	return nil, errors.New("storage not found")
}
