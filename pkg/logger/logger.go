package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// globalLogger is the package-level logger instance.
var globalLogger *zap.Logger

// Init initializes the global logger with the given configuration.
// It creates three zapcore.Core instances (console, app file, error file)
// and tees them together into a single logger.
func Init(cfg Config) error {
	cfg = cfg.withDefaults()
	level := cfg.parseLevel()

	// Console core: human-readable output to stdout, all levels.
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	// App file core: compact JSON, filtered by configured level.
	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "t",
		LevelKey:       "l",
		NameKey:        "n",
		CallerKey:      "c",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "m",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	appWriter := NewDailyRotateWriter(cfg.Dir, cfg.AppFilename)
	appCore := zapcore.NewCore(
		jsonEncoder,
		zapcore.AddSync(appWriter),
		level,
	)

	// Error file core: compact JSON, error level and above.
	errorWriter := NewDailyRotateWriter(cfg.Dir, cfg.ErrorFilename)
	errorCore := zapcore.NewCore(
		jsonEncoder,
		zapcore.AddSync(errorWriter),
		zapcore.ErrorLevel,
	)

	core := zapcore.NewTee(consoleCore, appCore, errorCore)
	globalLogger = zap.New(core)

	return nil
}

// Sync flushes any buffered log entries.
// It ignores "sync /dev/stdout: bad file descriptor" errors which occur
// on macOS when syncing stdout.
func Sync() error {
	if globalLogger == nil {
		return nil
	}
	err := globalLogger.Sync()
	if err == nil {
		return nil
	}
	// Ignore the known stdout sync error on macOS.
	if strings.Contains(err.Error(), "sync /dev/stdout: bad file descriptor") {
		return nil
	}
	return err
}

// Debug logs a message at DebugLevel.
func Debug(msg string, fields ...zap.Field) {
	if globalLogger != nil {
		globalLogger.Debug(msg, fields...)
	}
}

// Info logs a message at InfoLevel.
func Info(msg string, fields ...zap.Field) {
	if globalLogger != nil {
		globalLogger.Info(msg, fields...)
	}
}

// Warn logs a message at WarnLevel.
func Warn(msg string, fields ...zap.Field) {
	if globalLogger != nil {
		globalLogger.Warn(msg, fields...)
	}
}

// Error logs a message at ErrorLevel.
func Error(msg string, fields ...zap.Field) {
	if globalLogger != nil {
		globalLogger.Error(msg, fields...)
	}
}

// Fatal logs a message at FatalLevel, then calls os.Exit(1).
func Fatal(msg string, fields ...zap.Field) {
	if globalLogger != nil {
		globalLogger.Fatal(msg, fields...)
	}
}
