// Модуль shortner сокращает ссылки и перенаправляет пользователя на полную.
package shortner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"github.com/playmixer/short-link/internal/core/shortner"
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
		ctx    context.Context
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
				ctx:    context.Background(),
				userID: "1",
				link:   "https://practicum.yandex.ru/",
			},
			wantErr: nil,
		},
	}

	s := createStorage(t)
	sh := shortner.New(context.Background(), s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link, err := sh.Shorty(tt.args.ctx, tt.args.userID, tt.args.link)
			require.NoError(t, err)
			if tt.wantSLink != "" {
				require.Equal(t, tt.wantSLink, link)
			}
		})
	}
}

func TestShortner_PingStore(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
			},
			wantErr: nil,
		},
	}

	s := createStorage(t)
	sh := shortner.New(context.Background(), s)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sh.PingStore(tt.args.ctx)
			require.NoError(t, err)
		})
	}
}
