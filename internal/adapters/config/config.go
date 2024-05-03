package config

import (
	"flag"

	"github.com/playmixer/short-link/internal/adapters/api"
	"github.com/playmixer/short-link/internal/adapters/api/rest"
	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"github.com/playmixer/short-link/internal/core/shortner"
)

type Config struct {
	Api      api.Config
	Store    storage.Config
	Shortner shortner.Config
	BaseUrl  string
}

func Init() *Config {

	cfg := Config{
		Api:   api.Config{Rest: &rest.Config{}},
		Store: storage.Config{Memory: &memory.Config{}},
	}

	flag.StringVar(&cfg.Api.Rest.Addr, "a", "localhost:8080", "address listen")
	flag.StringVar(&cfg.BaseUrl, "b", "http://localhost:8080", "base url")

	flag.Parse()

	return &cfg
}
