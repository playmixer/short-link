package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/database"
	"github.com/playmixer/short-link/internal/adapters/storage/file"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"go.uber.org/zap"
)

type Config struct {
	Memory   *memory.Config
	File     *file.Config
	Database *database.Config
}

type Store interface {
	Set(ctx context.Context, key, value string) (string, error)
	Get(ctx context.Context, key string) (string, error)
	SetBatch(ctx context.Context, batch []models.ShortLink) ([]models.ShortLink, error)
	Ping(ctx context.Context) error
}

func NewStore(ctx context.Context, cfg *Config, log *zap.Logger) (Store, error) {
	if cfg.Database != nil && cfg.Database.DSN != "" {
		cfg.Database.SetLogger(log)
		store, err := database.New(ctx, cfg.Database)
		if err != nil {
			return nil, fmt.Errorf("failed initialize database storage: %w", err)
		}
		log.Info("database storage initialized")
		return store, nil
	}

	if cfg.File != nil && cfg.File.StoragePath != "" {
		store, err := file.New(cfg.File)
		if err != nil {
			return nil, fmt.Errorf("failed initialize file storage: %w", err)
		}
		log.Info("file storage initialized")
		return store, nil
	}

	if cfg.Memory != nil {
		store, err := memory.New(cfg.Memory)
		if err != nil {
			return nil, fmt.Errorf("failed initialize memory storage: %w", err)
		}
		log.Info("memory storage initialized")
		return store, nil
	}

	return nil, errors.New("storage not found")
}
