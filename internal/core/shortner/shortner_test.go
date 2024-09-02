// Модуль shortner сокращает ссылки и перенаправляет пользователя на полную.
package shortner

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
)

func createStorage(t *testing.T) storage.Store {
	t.Helper()

	s, err := storage.NewStore(context.Background(), &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestShortner_Shorty(t *testing.T) {
	type args struct {
		userID string
		link   string
	}
	tests := []struct {
		name      string
		args      args
		wantSLink string
		wantErr   error
	}{
		{
			name: "shorted",
			args: args{
				userID: "1",
				link:   "https://practicum.yandex.ru/",
			},
			wantErr: nil,
		},
	}

	s := createStorage(t)
	sh := New(context.Background(), s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link, err := sh.Shorty(context.Background(), tt.args.userID, tt.args.link)
			require.NoError(t, err)
			if tt.wantSLink != "" {
				require.Equal(t, tt.wantSLink, link)
			}
		})
	}
}

func TestShortner_PingStore(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "success",
			wantErr: nil,
		},
	}

	s := createStorage(t)
	sh := New(context.Background(), s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sh.PingStore(context.Background())
			require.NoError(t, err)
		})
	}
}

func TestSetLogger(t *testing.T) {
	s := &Shortner{}
	lgr := zap.NewNop()
	SetLogger(lgr)(s)

	if reflect.ValueOf(lgr).Pointer() != reflect.ValueOf(s.log).Pointer() {
		t.Fatal("logger not set")
	}
}
