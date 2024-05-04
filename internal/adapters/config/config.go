package config

import (
	"flag"
	"log"

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
}

func Init() *Config {

	cfg := Config{
		API:   api.Config{Rest: &rest.Config{}},
		Store: storage.Config{Memory: &memory.Config{}},
	}

	flag.StringVar(&cfg.API.Rest.Addr, "a", "localhost:8080", "address listen")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base url")
	flag.Parse()

	_ = godotenv.Load(".env")

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	return &cfg
}
