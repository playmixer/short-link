package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/playmixer/short-link/internal/adapters/api/rest"
	"github.com/playmixer/short-link/internal/adapters/config"
	"github.com/playmixer/short-link/internal/adapters/logger"
	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/core/shortner"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("failed initialize config: %v", err)
		return
	}

	lgr, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed initialize logger: %v", err)
		return
	}

	store, err := storage.NewStore(ctx, &cfg.Store, lgr)
	if err != nil {
		cancel()
		log.Fatalf("failed initialize storage: %v", err)
		return
	}

	short := shortner.New(store)
	srv := rest.New(
		short,
		rest.Addr(cfg.API.Rest.Addr),
		rest.BaseURL(cfg.BaseURL),
		rest.Logger(lgr),
	)
	if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
		cancel()
		log.Fatal(err)
	}
	cancel()
}
