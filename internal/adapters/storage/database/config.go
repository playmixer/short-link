package database

import "go.uber.org/zap"

type Config struct {
	log *zap.Logger
	DSN string `env:"DATABASE_DSN"`
}

func (c *Config) SetLogger(log *zap.Logger) {
	c.log = log
}
