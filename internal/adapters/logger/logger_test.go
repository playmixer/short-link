package logger_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/logger"
)

func getFileLogger(filename string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{filename}
	cfg.ErrorOutputPaths = []string{filename}

	lgr, _ := cfg.Build()
	return lgr
}

func BenchmarkLogger(b *testing.B) {
	lgr := getFileLogger("logger_output.log")
	defer func() { _ = lgr.Sync() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lgr.Info("This is a structured log to file.",
			zap.String("key", "value"),
			zap.Int("count", i),
		)
	}
}

func BenchmarkSugaredLogger(b *testing.B) {
	lgr := getFileLogger("sugaredlogger_output.log")
	sugar := lgr.Sugar()
	defer func() { _ = lgr.Sync() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Infof("This is a sugared log to file with key %s and count %d.", "value", i)
	}
}

func TestNew(t *testing.T) {
	_, err := logger.New("default")
	require.Error(t, err)
}
