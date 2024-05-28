package file

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/playmixer/short-link/internal/adapters/storage/memory"
)

type storeItem struct {
	ID          string `json:"id"`
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
			err = s.Store.Set(item.ShortURL, item.OriginalURL)
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

func (s *Store) Set(key, value string) error {
	if s.filepath != "" {
		f, err := os.OpenFile(s.filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed open file: %w", err)
		}
		item := storeItem{
			ID:          strconv.Itoa(time.Now().UTC().Nanosecond()),
			ShortURL:    key,
			OriginalURL: value,
		}
		b, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed marshal storage item: %w", err)
		}
		_, err = f.WriteString(string(b) + "\n")
		if err != nil {
			return fmt.Errorf("failed write to file storage: %w", err)
		}
	}

	err := s.Store.Set(key, value)
	if err != nil {
		return fmt.Errorf("failed setted data: %w", err)
	}

	return nil
}
