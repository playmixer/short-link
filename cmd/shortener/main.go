package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/api/rest"
	"github.com/playmixer/short-link/internal/adapters/config"
	"github.com/playmixer/short-link/internal/adapters/logger"
	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/core/shortner"
	"github.com/playmixer/short-link/pkg/util"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Println("Build verson: " + util.BuildData(buildVersion))
	fmt.Println("Build date: " + util.BuildData(buildDate))
	fmt.Println("Build commit: " + util.BuildData(buildCommit))
	if err := run(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Init()
	if err != nil {
		return fmt.Errorf("failed initialize config: %w", err)
	}

	lgr, err := logger.New(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed initialize logger: %w", err)
	}

	store, err := storage.NewStore(ctx, &cfg.Store, lgr)
	if err != nil {
		lgr.Error("failed initialize storage", zap.Error(err))
		return fmt.Errorf("failed initialize storage: %w", err)
	}

	short := shortner.New(ctx, store, shortner.SetLogger(lgr))
	srv := rest.New(
		short,
		rest.Addr(cfg.API.Rest.Addr),
		rest.BaseURL(cfg.BaseURL),
		rest.Logger(lgr),
		rest.SecretKey([]byte(cfg.API.Rest.SecretKey)),
	)
	err = srv.Run()
	if err != nil {
		return fmt.Errorf("stop server: %w", err)
	}
	return nil
}
