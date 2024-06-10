package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
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

func (s *Store) Set(ctx context.Context, key, value string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.data {
		if v == value {
			return k, storeerror.ErrNotUnique
		}
	}
	if _, ok := s.data[key]; ok {
		return key, storeerror.ErrDuplicateShortURL
	}
	s.data[key] = value
	return key, nil
}

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		return s.data[key], nil
	}
	return "", storeerror.ErrNotFoundKey
}

func (s *Store) SetBatch(ctx context.Context, batch []models.ShortLink) (output []models.ShortLink, err error) {
	for _, b := range batch {
		if _, err := s.Get(ctx, b.ShortURL); err == nil {
			return []models.ShortLink{}, storeerror.ErrDuplicateShortURL
		}
		if shortURL, err := s.GetByOriginal(ctx, b.OriginalURL); err == nil {
			return []models.ShortLink{{ShortURL: shortURL, OriginalURL: b.OriginalURL}}, storeerror.ErrNotUnique
		}
	}
	shortAppled := make([]string, 0)
	for _, req := range batch {
		_, err := s.Set(ctx, req.ShortURL, req.OriginalURL)
		if err != nil {
			if !errors.Is(err, storeerror.ErrDuplicateShortURL) {
				for _, a := range shortAppled {
					s.DeleteShortURL(ctx, a)
				}
			}
			return output, fmt.Errorf("set link `%s` failed: %w", req.OriginalURL, err)
		}
		shortAppled = append(shortAppled, req.ShortURL)
	}
	return output, nil
}

func (s *Store) GetByOriginal(ctx context.Context, original string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.data {
		if v == original {
			return k, nil
		}
	}
	return "", fmt.Errorf("not found short by original: %s", original)
}

func (s *Store) DeleteShortURL(ctx context.Context, short string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, short)
}

func (s *Store) Ping(ctx context.Context) error {
	return nil
}
