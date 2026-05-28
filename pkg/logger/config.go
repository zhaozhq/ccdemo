package logger

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

// Config holds the configuration for the logger.
type Config struct {
	Dir           string // log directory, default "./logs"
	AppFilename   string // normal log filename prefix, default "app"
	ErrorFilename string // error log filename prefix, default "error"
	FileMinLevel  string // minimum level for file output, default "info"
}

// withDefaults fills empty fields with default values.
func (c Config) withDefaults() Config {
	if c.Dir == "" {
		c.Dir = "./logs"
	}
	if c.AppFilename == "" {
		c.AppFilename = "app"
	}
	if c.ErrorFilename == "" {
		c.ErrorFilename = "error"
	}
	if c.FileMinLevel == "" {
		c.FileMinLevel = "info"
	}
	return c
}

// parseLevel parses the level string into a zapcore.Level.
// Supports: "debug", "info", "warn"/"warning", "error", "fatal", "panic" (case-insensitive).
// Invalid or empty values default to zapcore.InfoLevel.
func (c Config) parseLevel() zapcore.Level {
	level := strings.ToLower(strings.TrimSpace(c.FileMinLevel))
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}
