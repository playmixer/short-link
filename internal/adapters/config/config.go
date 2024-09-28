package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/api"
	"github.com/playmixer/short-link/internal/adapters/api/rest"
	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/adapters/storage/database"
	"github.com/playmixer/short-link/internal/adapters/storage/file"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"github.com/playmixer/short-link/internal/core/shortner"
)

// Config конфигурация сервиса.
type Config struct {
	API           api.Config
	Store         storage.Config
	Shortner      shortner.Config
	BaseURL       string `env:"BASE_URL"`
	LogLevel      string `env:"LOG_LEVEL"`
	ConfigPath    string `env:"CONFIG"`
	TrastedSubnet string `env:"TRUSTED_SUBNET"`
}

// Init инициализирует конфигурацию сервиса.
func Init() (*Config, error) {
	cfg := Config{
		API: api.Config{Rest: &rest.Config{}},
		Store: storage.Config{
			File:     &file.Config{},
			Database: &database.Config{},
			Memory:   &memory.Config{},
		},
	}
	cfg.Store.Database.SetLogger(zap.NewNop())

	flag.StringVar(&cfg.API.Rest.Addr, "a", "localhost:8080", "address listen")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&cfg.LogLevel, "l", "info", "logger level")
	flag.StringVar(&cfg.Store.File.StoragePath, "f", "", "storage file")
	flag.StringVar(&cfg.Store.Database.DSN, "d", "", "database dsn")
	flag.BoolVar(&cfg.API.Rest.HTTPSEnable, "s", false, "tls enable")
	flag.StringVar(&cfg.ConfigPath, "c", "", "file configuration")
	flag.StringVar(&cfg.ConfigPath, "config", "", "file configuration")
	flag.StringVar(&cfg.TrastedSubnet, "t", "", "trunsted subnet")
	flag.Parse()

	_ = godotenv.Load(".env")

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error parse config %w", err)
	}
	if cfg.ConfigPath != "" {
		err := fromFile(cfg.ConfigPath, &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed load configure from file: %w", err)
		}
	}

	return &cfg, nil
}

type configFile struct {
	ServerAddress   *string `json:"server_address"`
	BaseURL         *string `json:"base_url"`
	FileStoragePath *string `json:"file_storage_path"`
	DatabaseDSN     *string `json:"database_dsn"`
	EnableHTTPS     *bool   `json:"enable_https"`
	TrustedSubnet   *string `json:"trusted_subner"`
}

func fromFile(filepath string, cfg *Config) error {
	_, err := os.Stat(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file not exists from path `%s`", filepath)
		}
		return fmt.Errorf("failed get state of file: %w", err)
	}
	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed optn file: %w", err)
	}
	bBody, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed read file: %w", err)
	}

	configuration := configFile{}
	err = json.Unmarshal(bBody, &configuration)
	if err != nil {
		return fmt.Errorf("failed parse configuration: %w", err)
	}

	if configuration.BaseURL != nil && cfg.BaseURL == "" {
		cfg.BaseURL = *configuration.BaseURL
	}
	if configuration.DatabaseDSN != nil && cfg.Store.Database.DSN == "" {
		cfg.Store.Database.DSN = *configuration.DatabaseDSN
	}
	if configuration.EnableHTTPS != nil && !cfg.API.Rest.HTTPSEnable {
		cfg.API.Rest.HTTPSEnable = *configuration.EnableHTTPS
	}
	if configuration.FileStoragePath != nil && cfg.Store.File.StoragePath == "" {
		cfg.Store.File.StoragePath = *configuration.FileStoragePath
	}
	if configuration.ServerAddress != nil && cfg.API.Rest.Addr == "" {
		cfg.API.Rest.Addr = *configuration.ServerAddress
	}
	if configuration.TrustedSubnet != nil && cfg.TrastedSubnet == "" {
		cfg.TrastedSubnet = *configuration.TrustedSubnet
	}

	return nil
}
