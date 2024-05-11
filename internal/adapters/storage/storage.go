package storage

import (
	"errors"

	"github.com/playmixer/short-link/internal/adapters/storage/memory"
)

type Config struct {
	Memory *memory.Config
}

type Store interface {
	Set(key, value string) error
	Get(key string) (string, error)
}

func NewStore(cfg *Config) (Store, error) {
	if cfg.Memory != nil {
		return memory.New(cfg.Memory), nil
	}

	return nil, errors.New("storage not found")
}
