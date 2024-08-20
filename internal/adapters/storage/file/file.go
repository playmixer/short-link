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

// Store имлементация файлового хранилища.
type Store struct {
	*memory.Store
	filepath string
}

// New создает Store.
func New(cfg *Config) (*Store, error) {
	m, err := memory.New(&memory.Config{})
	if err != nil {
		return nil, fmt.Errorf("can`t initialize memory storage: %w", err)
	}
	s := &Store{
		Store:    m,
		filepath: cfg.StoragePath,
	}
	err = s.uploadFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed upload from file: %w", err)
	}

	return s, nil
}

// Set Сохраняет ссылку.
func (s *Store) Set(ctx context.Context, userID, key, value string) (string, error) {
	shortURL, err := s.Store.Set(ctx, userID, key, value)
	if err != nil {
		return shortURL, fmt.Errorf("failed setting data: %w", err)
	}

	if s.filepath != "" {
		f, err := os.OpenFile(s.filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err != nil {
			s.Store.RemoveShortURL(ctx, userID, shortURL)
			return "", fmt.Errorf("failed open file: %w", err)
		}
		item := memory.StoreItem{
			ID:          strconv.Itoa(time.Now().UTC().Nanosecond()),
			UserID:      userID,
			ShortURL:    shortURL,
			OriginalURL: value,
			IsDeleted:   false,
		}
		b, err := json.Marshal(item)
		if err != nil {
			s.Store.RemoveShortURL(ctx, userID, shortURL)
			return "", fmt.Errorf("failed marshal storage item: %w", err)
		}
		_, err = f.WriteString(string(b) + "\n")
		if err != nil {
			s.Store.RemoveShortURL(ctx, userID, shortURL)
			return "", fmt.Errorf("failed write to file storage: %w", err)
		}
	}

	return shortURL, nil
}

// SetBatch Сохраняет список ссылок.
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

// DeleteShortURLs Мягкое удаляет ссылки.
func (s *Store) DeleteShortURLs(ctx context.Context, shorts []models.ShortLink) error {
	err := s.Store.DeleteShortURLs(ctx, shorts)
	if err != nil {
		return fmt.Errorf("failed deleting shorts: %w", err)
	}
	err = s.reWriteStore()
	if err != nil {
		return fmt.Errorf("failed rewrite storage in file: %w", err)
	}

	return nil
}

func (s *Store) reWriteStore() error {
	f, err := os.OpenFile(s.filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed opend file: %w", err)
	}
	for _, v := range s.GetAll() {
		item := memory.StoreItem{
			ID:          v.ID,
			UserID:      v.UserID,
			ShortURL:    v.ShortURL,
			OriginalURL: v.OriginalURL,
			IsDeleted:   v.IsDeleted,
		}
		line, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed marshal data: %w", err)
		}
		_, err = f.WriteString(string(line) + "\n")
		if err != nil {
			return fmt.Errorf("failed write to file: %w", err)
		}
	}

	return nil
}

// HardDeleteURLs Хард удаление ссылок.
func (s *Store) HardDeleteURLs(ctx context.Context) error {
	err := s.Store.HardDeleteURLs(ctx)
	if err != nil {
		return fmt.Errorf("failed hard deleting URLs: %w", err)
	}
	err = s.reWriteStore()
	if err != nil {
		return fmt.Errorf("faile rewrite file store: %w", err)
	}

	return nil
}

func (s *Store) uploadFromFile() error {
	if s.filepath != "" {
		var f *os.File
		if _, err := os.Stat(s.filepath); errors.Is(err, os.ErrNotExist) {
			path := filepath.Dir(s.filepath)
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed create path %s storage: %w", path, err)
			}
			f, err = os.OpenFile(s.filepath, os.O_CREATE|os.O_RDONLY, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed create storage file: %w", err)
			}
			_, err = f.Seek(0, 0)
			if err != nil {
				return fmt.Errorf("failed seek file: %w", err)
			}
		} else {
			f, err = os.OpenFile(s.filepath, os.O_RDONLY, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed open storage file: %w", err)
			}
		}
		defer func() { _ = f.Close() }()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var item memory.StoreItem
			err := json.Unmarshal(scanner.Bytes(), &item)
			if err != nil {
				return fmt.Errorf("failed unmarshal data from storage: %w", err)
			}
			_, err = s.Store.Set(context.Background(), item.UserID, item.ShortURL, item.OriginalURL)
			if err != nil {
				return fmt.Errorf("failed set (%s, %s, %s): %w", item.UserID, item.ShortURL, item.OriginalURL, err)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed scanner file storage: %w", err)
		}
	}
	return nil
}
