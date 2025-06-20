package util

import (
	"log/slog"
	"os"
)

const (
	LogLevelError = "Error"
	LogLevelInfo  = "Info"
	LogLevelDebug = "Debug"
)

func Logger(logLevelEnv string) *slog.Logger {
	// Set log level based on environment
	var logLevel slog.Level

	switch logLevelEnv {
	case LogLevelError:
		logLevel = slog.LevelError
	case LogLevelInfo:
		logLevel = slog.LevelInfo
	case LogLevelDebug:
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelError
	}

	// Set up the logger with the chosen log level
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	return logger
}

// Logger returns a slog.Logger instance configured with the specified log level.
