package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected zapcore.Level
	}{
		{"debug lowercase", "debug", zapcore.DebugLevel},
		{"info lowercase", "info", zapcore.InfoLevel},
		{"warn lowercase", "warn", zapcore.WarnLevel},
		{"warning lowercase", "warning", zapcore.WarnLevel},
		{"error lowercase", "error", zapcore.ErrorLevel},
		{"fatal lowercase", "fatal", zapcore.FatalLevel},
		{"panic lowercase", "panic", zapcore.PanicLevel},

		{"debug uppercase", "DEBUG", zapcore.DebugLevel},
		{"info uppercase", "INFO", zapcore.InfoLevel},
		{"warn uppercase", "WARN", zapcore.WarnLevel},
		{"warning uppercase", "WARNING", zapcore.WarnLevel},
		{"error uppercase", "ERROR", zapcore.ErrorLevel},
		{"fatal uppercase", "FATAL", zapcore.FatalLevel},
		{"panic uppercase", "PANIC", zapcore.PanicLevel},

		{"debug mixed case", "DeBuG", zapcore.DebugLevel},
		{"info mixed case", "InFo", zapcore.InfoLevel},
		{"warn mixed case", "WaRn", zapcore.WarnLevel},
		{"warning mixed case", "WaRnInG", zapcore.WarnLevel},

		{"invalid level", "invalid", zapcore.InfoLevel},
		{"empty level", "", zapcore.InfoLevel},
		{"whitespace only", "   ", zapcore.InfoLevel},
		{"unknown level", "trace", zapcore.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{FileMinLevel: tt.level}
			got := c.parseLevel()
			if got != tt.expected {
				t.Errorf("parseLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWithDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    Config
		expected Config
	}{
		{
			name:     "all empty fields",
			input:    Config{},
			expected: Config{Dir: "./logs", AppFilename: "app", ErrorFilename: "error", FileMinLevel: "info"},
		},
		{
			name:     "some fields set",
			input:    Config{Dir: "/var/log", AppFilename: "myapp"},
			expected: Config{Dir: "/var/log", AppFilename: "myapp", ErrorFilename: "error", FileMinLevel: "info"},
		},
		{
			name:     "all fields set",
			input:    Config{Dir: "/var/log", AppFilename: "myapp", ErrorFilename: "err", FileMinLevel: "debug"},
			expected: Config{Dir: "/var/log", AppFilename: "myapp", ErrorFilename: "err", FileMinLevel: "debug"},
		},
		{
			name:     "only FileMinLevel empty",
			input:    Config{Dir: "/var/log", AppFilename: "myapp", ErrorFilename: "err"},
			expected: Config{Dir: "/var/log", AppFilename: "myapp", ErrorFilename: "err", FileMinLevel: "info"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.withDefaults()
			if got != tt.expected {
				t.Errorf("withDefaults() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}
