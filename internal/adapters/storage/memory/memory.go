package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNotFoundKey = errors.New("not found value by key")
)

type Store struct {
	data map[string]string
	mu   *sync.Mutex
}

func New(cfg *Config) (*Store, error) {
	return &Store{
		data: make(map[string]string),
		mu:   &sync.Mutex{},
	}, nil
}

func (s *Store) Set(ctx context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		return fmt.Errorf("short link `%s` is exists", key)
	}
	s.data[key] = value
	return nil
}

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		return s.data[key], nil
	}
	return "", ErrNotFoundKey
}
