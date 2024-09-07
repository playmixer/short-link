package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/adapters/storage/database"
	"github.com/playmixer/short-link/internal/adapters/storage/file"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
)

func removeFileStorage(t *testing.T) {
	t.Helper()
	err := os.Remove("./data.json")
	if err != nil {
		t.Fatal(err)
	}
}
func TestNewStore(t *testing.T) {
	defer removeFileStorage(t)
	type args struct {
		cfg *storage.Config
		log *zap.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    storage.Store
		wantErr error
	}{
		{
			name: "memory",
			args: args{
				cfg: &storage.Config{Memory: &memory.Config{}},
				log: zap.NewNop(),
			},
			want:    &memory.Store{},
			wantErr: nil,
		},
		{
			name: "file",
			args: args{
				cfg: &storage.Config{File: &file.Config{StoragePath: "./data.json"}},
				log: zap.NewNop(),
			},
			want:    &file.Store{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.NewStore(context.Background(), tt.args.cfg, tt.args.log)
			require.NoError(t, err)
			require.IsType(t, got, tt.want)
		})
	}
}

func TestNewStoreErr(t *testing.T) {
	type args struct {
		cfg *storage.Config
		log *zap.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    storage.Store
		wantErr error
	}{
		{
			name: "database",
			args: args{
				cfg: &storage.Config{Database: &database.Config{DSN: "file:test.db?cache=shared&mode=memory"}},
				log: zap.NewNop(),
			},
			want:    nil,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.NewStore(context.Background(), tt.args.cfg, tt.args.log)
			require.Error(t, err)
			require.IsType(t, got, tt.want)
		})
	}
}
