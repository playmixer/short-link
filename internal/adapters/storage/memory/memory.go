package memory

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
)

// StoreItem элемент хранения ссылки.
type StoreItem struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	IsDeleted   bool   `json:"is_deleted"`
}

// Store имплементация хранилища.
type Store struct {
	mu   *sync.Mutex
	data []StoreItem
}

// New создает Store.
func New(cfg *Config) (*Store, error) {
	return &Store{
		data: make([]StoreItem, 0),
		mu:   &sync.Mutex{},
	}, nil
}

func (s *Store) Close() {}

// Set Сохраняет ссылку.
func (s *Store) Set(ctx context.Context, userID, shortURL, originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.data {
		if v.OriginalURL == originalURL && v.UserID == userID {
			return v.ShortURL, storeerror.ErrNotUnique
		}
		if v.ShortURL == shortURL && v.UserID == userID {
			return v.ShortURL, storeerror.ErrDuplicateShortURL
		}
	}

	s.data = append(s.data, StoreItem{
		ID:          strconv.Itoa(time.Now().Nanosecond()),
		UserID:      userID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})

	return shortURL, nil
}

// GetByUser Возвращает оригинальную ссылку пользователя.
func (s *Store) GetByUser(ctx context.Context, userID, shortURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.data {
		if v.ShortURL == shortURL && v.UserID == userID {
			return v.OriginalURL, nil
		}
	}
	return "", storeerror.ErrNotFoundKey
}

// Get Возвращает оригинальную ссылку.
func (s *Store) Get(ctx context.Context, shortURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.data {
		if v.ShortURL == shortURL {
			if v.IsDeleted {
				return v.OriginalURL, storeerror.ErrShortURLDeleted
			}
			return v.OriginalURL, nil
		}
	}
	return "", storeerror.ErrNotFoundKey
}

// SetBatch Сохраняет список ссылок.
func (s *Store) SetBatch(ctx context.Context, userID string, batch []models.ShortLink) (
	output []models.ShortLink,
	err error,
) {
	for _, b := range batch {
		if _, err := s.GetByUser(ctx, userID, b.ShortURL); err == nil {
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
					s.RemoveShortURL(ctx, userID, a)
				}
			}
			return []models.ShortLink{}, fmt.Errorf("set link `%s` failed: %w", req.OriginalURL, err)
		}
		shortAppled = append(shortAppled, req.ShortURL)
		output = append(output, req)
	}
	return output, nil
}

// GetByOriginal возврашает коротку ссылку по оригинальной.
func (s *Store) GetByOriginal(ctx context.Context, userID, originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.data {
		if v.OriginalURL == originalURL && v.UserID == userID {
			return v.ShortURL, nil
		}
	}
	return "", fmt.Errorf("not found short by original URL: %s", originalURL)
}

// RemoveShortURL удаление ссылки.
func (s *Store) RemoveShortURL(ctx context.Context, userID, short string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	newStorage := []StoreItem{}
	for _, v := range s.data {
		if !(v.ShortURL == short && v.UserID == userID) {
			newStorage = append(newStorage, v)
		}
	}
	s.data = newStorage
}

// Ping Проверка соединения с хранилищем.
func (s *Store) Ping(ctx context.Context) error {
	return nil
}

// GetAllURL Возвращает все ссылки пользователя.
func (s *Store) GetAllURL(ctx context.Context, userID string) ([]models.ShortenURL, error) {
	result := []models.ShortenURL{}
	for _, v := range s.data {
		if v.UserID == userID {
			result = append(result, models.ShortenURL{ShortURL: v.ShortURL, OriginalURL: v.OriginalURL})
		}
	}
	return result, nil
}

// GetAll возвращает все ссылки.
func (s *Store) GetAll() []StoreItem {
	return s.data
}

// DeleteShortURLs Мягкое удаляет ссылки.
func (s *Store) DeleteShortURLs(ctx context.Context, shorts []models.ShortLink) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, short := range shorts {
		for i, v := range s.data {
			if v.ShortURL == short.ShortURL && v.UserID == short.UserID {
				s.data[i].IsDeleted = true
				break
			}
		}
	}
	return nil
}

// HardDeleteURLs Хард удаление ссылок.
func (s *Store) HardDeleteURLs(ctx context.Context) error {
	newData := make([]StoreItem, 0)
	for _, v := range s.data {
		if !v.IsDeleted {
			newData = append(newData, v)
		}
	}
	s.data = newData

	return nil
}
