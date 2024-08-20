package logger_test

import (
	"testing"

	"go.uber.org/zap"
)

func getFileLogger(filename string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{filename}
	cfg.ErrorOutputPaths = []string{filename}

	logger, _ := cfg.Build()
	return logger
}

func BenchmarkLoggerFile(b *testing.B) {
	logger := getFileLogger("logger_output.log")
	defer logger.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("This is a structured log to file.",
			zap.String("key", "value"),
			zap.Int("count", i),
		)
	}
}

func BenchmarkSugaredLoggerFile(b *testing.B) {
	logger := getFileLogger("sugaredlogger_output.log")
	sugar := logger.Sugar()
	defer sugar.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Infof("This is a sugared log to file with key %s and count %d.", "value", i)
	}
}
