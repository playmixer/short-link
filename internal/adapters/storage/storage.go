package storage

import (
	"fmt"

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

	return nil, fmt.Errorf("storage not found")
}
