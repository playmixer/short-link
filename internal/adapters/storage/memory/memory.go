package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
)

type storageURL struct {
	userID      string
	ShortURL    string
	OriginalURL string
}

type Store struct {
	mu   *sync.Mutex
	data []storageURL
}

func New(cfg *Config) (*Store, error) {
	return &Store{
		data: make([]storageURL, 0),
		mu:   &sync.Mutex{},
	}, nil
}

func (s *Store) Set(ctx context.Context, userID, shortURL, originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.data {
		if v.OriginalURL == originalURL && v.userID == userID {
			return v.ShortURL, storeerror.ErrNotUnique
		}
		if v.ShortURL == shortURL && v.userID == userID {
			return v.ShortURL, storeerror.ErrDuplicateShortURL
		}
	}

	s.data = append(s.data, storageURL{
		userID:      userID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})

	return shortURL, nil
}

func (s *Store) Get(ctx context.Context, userID, shortURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.data {
		if v.ShortURL == shortURL && v.userID == userID {
			return v.OriginalURL, nil
		}
	}
	return "", storeerror.ErrNotFoundKey
}

func (s *Store) SetBatch(ctx context.Context, userID string, batch []models.ShortLink) (
	output []models.ShortLink,
	err error,
) {
	for _, b := range batch {
		if _, err := s.Get(ctx, userID, b.ShortURL); err == nil {
			return []models.ShortLink{}, storeerror.ErrDuplicateShortURL
		}
		if shortURL, err := s.GetByOriginal(ctx, userID, b.OriginalURL); err == nil {
			return []models.ShortLink{{ShortURL: shortURL, OriginalURL: b.OriginalURL}}, storeerror.ErrNotUnique
		}
	}
	shortAppled := make([]string, 0)
	for _, req := range batch {
		_, err := s.Set(ctx, userID, req.ShortURL, req.OriginalURL)
		if err != nil {
			if !errors.Is(err, storeerror.ErrDuplicateShortURL) {
				for _, a := range shortAppled {
					s.DeleteShortURL(ctx, userID, a)
				}
			}
			return []models.ShortLink{}, fmt.Errorf("set link `%s` failed: %w", req.OriginalURL, err)
		}
		shortAppled = append(shortAppled, req.ShortURL)
		output = append(output, req)
	}
	return output, nil
}

func (s *Store) GetByOriginal(ctx context.Context, userID, originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.data {
		if v.OriginalURL == originalURL {
			return v.ShortURL, nil
		}
	}
	return "", fmt.Errorf("not found short by original URL: %s", originalURL)
}

func (s *Store) DeleteShortURL(ctx context.Context, userID, short string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	newStorage := []storageURL{}
	for _, v := range s.data {
		if !(v.ShortURL == short && v.userID == userID) {
			newStorage = append(newStorage, v)
		}
	}
	s.data = newStorage
}

func (s *Store) Ping(ctx context.Context) error {
	return nil
}

func (s *Store) GetAllURL(ctx context.Context, userID string) ([]models.ShortenURL, error) {
	result := []models.ShortenURL{}
	for _, v := range s.data {
		if v.userID == userID {
			result = append(result, models.ShortenURL{ShortURL: v.ShortURL, OriginalURL: v.OriginalURL})
		}
	}
	return result, nil
}
