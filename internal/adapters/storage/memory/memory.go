package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/shortnererror"
)

var (
	ErrNotFoundKey   = errors.New("not found value by key")
	ErrNotFoundValue = errors.New("not found value by value")
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
	for _, v := range s.data {
		if v == value {
			return shortnererror.ErrNotUnique
		}
	}
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

func (s *Store) SetBatch(ctx context.Context, batch []models.ShortLink) error {
	for _, req := range batch {
		err := s.Set(ctx, req.ShortURL, req.OriginalURL)
		if err != nil {
			return fmt.Errorf("set link `%s` failed: %w", req.OriginalURL, err)
		}
	}
	return nil
}

func (s *Store) GetByOriginal(ctx context.Context, original string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.data {
		if v == original {
			return k, nil
		}
	}
	return "", ErrNotFoundValue
}
