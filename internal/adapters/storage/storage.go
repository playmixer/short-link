package storage

import (
	"errors"
	"fmt"

	"github.com/playmixer/short-link/internal/adapters/storage/file"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
)

type Config struct {
	Memory *memory.Config
	File   *file.Config
}

type Store interface {
	Set(key, value string) error
	Get(key string) (string, error)
}

func NewStore(cfg *Config) (Store, error) {
	if cfg.Memory != nil {
		store, err := memory.New(cfg.Memory)
		if err != nil {
			return nil, fmt.Errorf("failed initialize memory storage: %w", err)
		}
		return store, nil
	}

	if cfg.File != nil {
		store, err := file.New(cfg.File)
		if err != nil {
			return nil, fmt.Errorf("failed initialize file storage: %w", err)
		}
		return store, nil
	}

	return nil, errors.New("storage not found")
}
