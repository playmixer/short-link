package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/playmixer/short-link/internal/adapters/api"
	"github.com/playmixer/short-link/internal/adapters/api/rest"
	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"github.com/playmixer/short-link/internal/core/shortner"
)

type Config struct {
	API      api.Config
	Store    storage.Config
	Shortner shortner.Config
	BaseURL  string `env:"BASE_URL"`
	LogLevel string `env:"LOG_LEVEL"`
}

func Init() (*Config, error) {
	cfg := Config{
		API:   api.Config{Rest: &rest.Config{}},
		Store: storage.Config{Memory: &memory.Config{}},
	}

	flag.StringVar(&cfg.API.Rest.Addr, "a", "localhost:8080", "address listen")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&cfg.LogLevel, "l", "info", "logger level")
	flag.Parse()

	_ = godotenv.Load(".env")

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error parse config %w", err)
	}

	return &cfg, nil
}
