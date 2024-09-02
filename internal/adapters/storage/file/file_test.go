package file_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/file"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
)

func createFileStorage(t *testing.T) *file.Store {
	t.Helper()

	s, err := file.New(&file.Config{
		StoragePath: "./data.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func removeFileStorage(t *testing.T) {
	t.Helper()
	err := os.Remove("./data.json")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStorage_Get(t *testing.T) {
	type cases struct {
		name     string
		short    string
		original string
		err      error
	}
	tests := []cases{
		{
			name:     "empty",
			short:    "WQEAWE",
			original: "",
			err:      storeerror.ErrNotFoundKey,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			ctx := context.Background()
			s := createFileStorage(tt)
			original, err := s.Get(ctx, test.short)
			require.Error(tt, err, storeerror.ErrNotFoundKey)
			require.Equal(tt, original, test.original)
		})
	}
	removeFileStorage(t)
}

func TestStorage_SetBatch(t *testing.T) {
	type cases struct {
		name   string
		userID string
		batch  []models.ShortLink
		want   []models.ShortLink
	}
	tests := []cases{
		{
			name:   "default",
			userID: "1",
			batch: []models.ShortLink{
				{ShortURL: "QWE123", OriginalURL: "https://stackoverflow.com/"},
			},
			want: []models.ShortLink{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			ctx := context.Background()
			s := createFileStorage(tt)
			res, err := s.SetBatch(ctx, test.userID, test.batch)
			require.NoError(tt, err)
			for _, r := range res {
				require.NotEmpty(tt, r.OriginalURL)
				require.NotEmpty(tt, r.ShortURL)
			}
		})
	}
	removeFileStorage(t)
}

func TestStorage_GetAll(t *testing.T) {
	type cases struct {
		name string
		res  []memory.StoreItem
		err  error
	}
	tests := []cases{
		{
			name: "empty",
			res:  []memory.StoreItem{},
			err:  nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			s := createFileStorage(tt)
			data := s.GetAll()
			require.Equal(tt, data, test.res)
		})
	}

	removeFileStorage(t)
}

func TestStorage_RemoveShortURL(t *testing.T) {
	shortLink := "EQWEe"
	type cases struct {
		name   string
		userID string
		short  string
		err    error
	}
	tests := []cases{
		{
			name:   "deleting short",
			userID: "1",
			err:    nil,
			short:  shortLink,
		},
	}
	s := createFileStorage(t)
	_, err := s.Set(context.Background(), "1", shortLink, "https://practicum.yandex.ru/")
	require.NoError(t, err)
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			ctx := context.Background()
			_, err := s.Get(ctx, test.short)
			require.NoError(t, err)
			s.RemoveShortURL(ctx, test.userID, test.short)
			for _, short := range s.GetAll() {
				if short.ShortURL == test.short && short.IsDeleted == false {
					t.Fatal("short is not deleted")
				}
			}
		})
	}
	removeFileStorage(t)
}

func TestStorage_HardDeleteURLs(t *testing.T) {
	shortLink := "EQWEe"
	type cases struct {
		name   string
		userID string
		short  string
		err    error
	}
	tests := []cases{
		{
			name:   "deleting short",
			userID: "1",
			err:    nil,
			short:  shortLink,
		},
	}
	s := createFileStorage(t)
	_, err := s.Set(context.Background(), "1", shortLink, "https://practicum.yandex.ru/")
	require.NoError(t, err)
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			ctx := context.Background()
			_, err := s.Get(ctx, test.short)
			require.NoError(t, err)
			s.RemoveShortURL(ctx, test.userID, test.short)
			err = s.HardDeleteURLs(ctx)
			require.NoError(t, err)
			for _, short := range s.GetAll() {
				if short.ShortURL == test.short {
					t.Fatal("short is not deleted")
				}
			}
		})
	}
	removeFileStorage(t)
}

func TestStorage_Ping(t *testing.T) {
	ctx := context.Background()
	s := createFileStorage(t)
	err := s.Ping(ctx)
	require.NoError(t, err)
}
