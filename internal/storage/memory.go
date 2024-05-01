package storage

import (
	"sync"
)

type Store struct {
	data map[string]string
	mu   sync.Mutex
}

func New() *Store {
	return &Store{
		data: make(map[string]string),
		mu:   sync.Mutex{},
	}
}

func (s *Store) Add(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Store) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		return s.data[key], nil
	}
	return "", ErrNotFoundKey
}
