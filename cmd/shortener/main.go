package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/api"
	"github.com/playmixer/short-link/internal/adapters/auth"
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

	shutdownDelay = time.Second * 2
)

func main() {
	fmt.Println("Build verson: " + util.BuildData(buildVersion))
	fmt.Println("Build date: " + util.BuildData(buildDate))
	fmt.Println("Build commit: " + util.BuildData(buildCommit))
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

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

	authManager, err := auth.New(auth.SetLogger(lgr), auth.SetSecretKey([]byte(cfg.API.SecretKey)))
	if err != nil {
		return fmt.Errorf("failed initializa auth manager: %w", err)
	}

	short := shortner.New(ctx, store, shortner.SetLogger(lgr))
	srv, err := api.New(short, authManager, lgr, cfg.API)
	if err != nil {
		return fmt.Errorf("failed initialize api: %w", err)
	}

	lgr.Info("Starting")
	go func() {
		err = srv.Run()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			lgr.Error("stop server", zap.Error(err))
		}
	}()
	<-ctx.Done()
	lgr.Info("Stopping...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), shutdownDelay)
	defer cancel()

	srv.Stop()    // отключаем api сервер.
	short.Wait()  // ждем завершения горитин.
	store.Close() // закрываем соединение с бд.

	<-ctxShutdown.Done()
	lgr.Info("Service stoped")
	return nil
}
