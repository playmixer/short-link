package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// New создает логер.
func New(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("logger failed parse level %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed build logger %w", err)
	}

	return zl, nil
}
