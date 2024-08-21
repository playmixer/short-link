package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
)

func createMemoryStorage(t *testing.T) *memory.Store {
	t.Helper()

	s, err := memory.New(&memory.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return s
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
			s := createMemoryStorage(tt)
			original, err := s.Get(ctx, test.short)
			require.Error(tt, err, storeerror.ErrNotFoundKey)
			require.Equal(tt, original, test.original)
		})
	}

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
			s := createMemoryStorage(tt)
			data := s.GetAll()
			require.Equal(tt, data, test.res)
		})
	}

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
	s := createMemoryStorage(t)
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
	s := createMemoryStorage(t)
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

}

func TestStorage_Ping(t *testing.T) {
	ctx := context.Background()
	s := createMemoryStorage(t)
	err := s.Ping(ctx)
	require.NoError(t, err)
}
