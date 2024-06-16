package file

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
)

type storeItem struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Store struct {
	*memory.Store
	filepath string
}

func New(cfg *Config) (*Store, error) {
	var f *os.File
	m, err := memory.New(&memory.Config{})
	if err != nil {
		return nil, fmt.Errorf("can`t initialize memory storage: %w", err)
	}
	s := &Store{
		Store:    m,
		filepath: cfg.StoragePath,
	}
	if s.filepath != "" {
		if _, err := os.Stat(s.filepath); errors.Is(err, os.ErrNotExist) {
			path := filepath.Dir(s.filepath)
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("failed create path %s storage: %w", path, err)
			}
			f, err = os.OpenFile(s.filepath, os.O_CREATE|os.O_RDONLY, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("failed create storage file: %w", err)
			}
			_, err = f.Seek(0, 0)
			if err != nil {
				return nil, fmt.Errorf("failed seek file: %w", err)
			}
		} else {
			f, err = os.OpenFile(s.filepath, os.O_RDONLY, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("failed open storage file: %w", err)
			}
		}
		defer func() { _ = f.Close() }()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var item storeItem
			err := json.Unmarshal(scanner.Bytes(), &item)
			if err != nil {
				return nil, fmt.Errorf("failed unmarshal data from storage: %w", err)
			}
			_, err = s.Store.Set(context.Background(), item.UserID, item.ShortURL, item.OriginalURL)
			if err != nil {
				return nil, fmt.Errorf("failed set: %w", err)
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed scanner file storage: %w", err)
		}
	}

	return s, nil
}

func (s *Store) Set(ctx context.Context, userID, key, value string) (string, error) {
	shortURL, err := s.Store.Set(ctx, userID, key, value)
	if err != nil {
		return shortURL, fmt.Errorf("failed setting data: %w", err)
	}

	if s.filepath != "" {
		f, err := os.OpenFile(s.filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err != nil {
			s.Store.DeleteShortURL(ctx, userID, shortURL)
			return "", fmt.Errorf("failed open file: %w", err)
		}
		item := storeItem{
			ID:          strconv.Itoa(time.Now().UTC().Nanosecond()),
			UserID:      userID,
			ShortURL:    shortURL,
			OriginalURL: value,
		}
		b, err := json.Marshal(item)
		if err != nil {
			s.Store.DeleteShortURL(ctx, userID, shortURL)
			return "", fmt.Errorf("failed marshal storage item: %w", err)
		}
		_, err = f.WriteString(string(b) + "\n")
		if err != nil {
			s.Store.DeleteShortURL(ctx, userID, shortURL)
			return "", fmt.Errorf("failed write to file storage: %w", err)
		}
	}

	return shortURL, nil
}

func (s *Store) SetBatch(ctx context.Context, userID string, batch []models.ShortLink) (
	output []models.ShortLink,
	err error,
) {
	for _, b := range batch {
		if _, err := s.Store.GetByUser(ctx, userID, b.ShortURL); err == nil {
			return []models.ShortLink{}, storeerror.ErrDuplicateShortURL
		}
		if shortURL, err := s.Store.GetByOriginal(ctx, userID, b.OriginalURL); err == nil {
			return []models.ShortLink{{ShortURL: shortURL, OriginalURL: b.OriginalURL}}, storeerror.ErrNotUnique
		}
	}

	for _, req := range batch {
		_, err := s.Set(ctx, userID, req.ShortURL, req.OriginalURL)
		if err != nil {
			return []models.ShortLink{}, fmt.Errorf("failed save data: %w", err)
		}
		output = append(output, req)
	}
	return output, nil
}
