package database

import "go.uber.org/zap"

// Config конфигурация подключени к базе данных.
type Config struct {
	log *zap.Logger
	DSN string `env:"DATABASE_DSN"`
}

// SetLogger установить логгер.
func (c *Config) SetLogger(log *zap.Logger) {
	c.log = log
}
