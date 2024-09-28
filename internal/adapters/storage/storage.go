package storage

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/database"
	"github.com/playmixer/short-link/internal/adapters/storage/file"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
)

// Config - конфигурация хранилища.
// Хранилище создастся того типа в зависимости какой параметр будет передан.
type Config struct {
	Memory   *memory.Config   // Memory - сохранение в памяти.
	File     *file.Config     // File - хранение в файле.
	Database *database.Config // Database - хранение в базе данных.
}

// Store - интерефейс хранилища ссылок.
type Store interface {
	// Возвращает оригинальную ссылку.
	Get(ctx context.Context, short string) (string, error)
	// Возвращает все ссылки пользователя.
	GetAllURL(ctx context.Context, userID string) ([]models.ShortenURL, error)
	// Сохраняет ссылку.
	Set(ctx context.Context, userID string, short string, url string) (string, error)
	// Сохраняет список ссылок.
	SetBatch(ctx context.Context, userID string, batch []models.ShortLink) ([]models.ShortLink, error)
	// Проверка соединения с хранилищем.
	Ping(ctx context.Context) error
	// Мягкое удаляет ссылки.
	DeleteShortURLs(ctx context.Context, shorts []models.ShortLink) error
	GetState(ctx context.Context) (urls int, users int, err error)
	// Хард удаление ссылок.
	HardDeleteURLs(ctx context.Context) error
	Close()
}

// NewStore - Создает хранилище.
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
