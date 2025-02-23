package logger

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var defaultLogger *zap.Logger

func init() {
	var err error
	defaultLogger, err = newLoggerWithConfig()
	if err != nil {
		panic(err)
	}
}

// Extract contextからロガーを取得
func Extract(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return defaultLogger
	}
	return defaultLogger
}

func DefaultLogger() *zap.Logger {
	return defaultLogger
}

func newLoggerWithConfig() (*zap.Logger, error) {
	var config zap.Config
	if os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "staging" {
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		config.Encoding = "json"
	} else {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.Development = true
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.Encoding = "console"
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	return config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
}
