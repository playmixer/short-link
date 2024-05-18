package main

import (
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
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("failed initialize config: %v", err)
		return
	}

	if err := logger.Init(cfg.LogLevel); err != nil {
		log.Fatalf("failed initialize logger: %v", err)
		return
	}

	store, err := storage.NewStore(&cfg.Store)
	if err != nil {
		log.Fatalf("failed initialize storage: %v", err)
		return
	}
	short := shortner.New(store)

	srv := rest.New(
		short,
		rest.Addr(cfg.API.Rest.Addr),
		rest.BaseURL(cfg.BaseURL),
	)
	if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
