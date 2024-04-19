package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

func Setup() (*zap.Logger, error) {
	logLevel := zapcore.InfoLevel

	envLogLevel := os.Getenv("LOG_LEVEL")
	if envLogLevel != "" {
		err := logLevel.Set(strings.ToLower(envLogLevel))
		if err != nil {
			return nil, err
		}
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(logLevel)

	return config.Build()
}
