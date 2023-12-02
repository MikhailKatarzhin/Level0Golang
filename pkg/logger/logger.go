// Package logger provides helper functions for using zap.Logger.
package logger

import (
	"fmt"
	"log"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logLevelEnvName = "LOG_LEVEL"

var (
	errr    error
	loggerr *zap.Logger
	once    sync.Once
)

func GetLogger() (*zap.Logger, error) {
	once.Do(func() {
		loggerr, errr = createLogger(os.Getenv(logLevelEnvName))
	})

	return loggerr, errr
}

var L = MustGetLogger

func MustGetLogger() *zap.Logger {
	logger, err := GetLogger()
	if err != nil {
		panic(err)
	}

	return logger
}

func Close() {
	if err := loggerr.Sync(); err != nil {
		log.Printf("can not to flush logger: %s", err.Error())
	}
}

func createLogger(logLevel string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.NameKey = "name"
	cfg.EncoderConfig.StacktraceKey = "stacktrace"
	cfg.EncoderConfig.TimeKey = "timestamp"

	if logLevel != "" {
		expectedLevel, err := zapcore.ParseLevel(logLevel)
		if err != nil {
			return nil, fmt.Errorf(
				"can not to parse LOG_LEVEL %q, error: %w",
				logLevel, err,
			)
		}

		cfg.Level = zap.NewAtomicLevelAt(expectedLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf(
			"can not to build config for zap logger, error: %w", err,
		)
	}

	return logger, nil
}
