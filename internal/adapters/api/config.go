package api

import (
	"github.com/playmixer/short-link/internal/adapters/api/grpch"
	"github.com/playmixer/short-link/internal/adapters/api/rest"
)

// Config is the configuration for the API.
type Config struct {
	Rest          *rest.Config
	GRPC          *grpch.Config
	SecretKey     string `env:"SECRET_KEY"`
	BaseURL       string `env:"BASE_URL"`
	TrustedSubnet string `env:"TRUSTED_SUBNET"`
}
