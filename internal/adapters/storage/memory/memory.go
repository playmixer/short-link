package memory

import (
	"fmt"
	"sync"
)

var (
	ErrNotFoundKey = fmt.Errorf("not found value by key")
)

type Store struct {
	data map[string]string
	mu   sync.Mutex
}

func New(cfg *Config) *Store {
	return &Store{
		data: make(map[string]string),
		mu:   sync.Mutex{},
	}
}

func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *Store) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		return s.data[key], nil
	}
	return "", ErrNotFoundKey
}