package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readLogLines reads all lines from a log file in the given directory
// matching the provided prefix.
func readLogLines(t *testing.T, dir, prefix string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read log dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix+".") && strings.HasSuffix(name, ".log") {
			f, err := os.Open(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("open log file: %v", err)
			}
			defer f.Close()

			var lines []string
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				t.Fatalf("scan log file: %v", err)
			}
			return lines
		}
	}
	return nil
}

// countMsgOccurrences counts how many log lines contain the given message.
func countMsgOccurrences(lines []string, msg string) int {
	count := 0
	for _, line := range lines {
		if strings.Contains(line, msg) {
			count++
		}
	}
	return count
}

func TestLoggerInit(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		Dir:           dir,
		AppFilename:   "app",
		ErrorFilename: "error",
		FileMinLevel:  "info",
	}

	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Sync()

	Info("info message")
	Error("error message")

	if err := Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	appLines := readLogLines(t, dir, "app")
	if len(appLines) != 2 {
		t.Fatalf("expected 2 app log lines, got %d", len(appLines))
	}
	if countMsgOccurrences(appLines, "info message") != 1 {
		t.Fatalf("expected 1 info message in app log")
	}
	if countMsgOccurrences(appLines, "error message") != 1 {
		t.Fatalf("expected 1 error message in app log")
	}

	errorLines := readLogLines(t, dir, "error")
	if len(errorLines) != 1 {
		t.Fatalf("expected 1 error log line, got %d", len(errorLines))
	}
	if countMsgOccurrences(errorLines, "error message") != 1 {
		t.Fatalf("expected 1 error message in error log")
	}
	if countMsgOccurrences(errorLines, "info message") != 0 {
		t.Fatalf("expected 0 info messages in error log")
	}
}

func TestLoggerLevelFilter(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		Dir:           dir,
		AppFilename:   "app",
		ErrorFilename: "error",
		FileMinLevel:  "warn",
	}

	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Sync()

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	if err := Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	appLines := readLogLines(t, dir, "app")
	if countMsgOccurrences(appLines, "debug message") != 0 {
		t.Fatalf("expected 0 debug messages in app log")
	}
	if countMsgOccurrences(appLines, "info message") != 0 {
		t.Fatalf("expected 0 info messages in app log")
	}
	if countMsgOccurrences(appLines, "warn message") != 1 {
		t.Fatalf("expected 1 warn message in app log")
	}
	if countMsgOccurrences(appLines, "error message") != 1 {
		t.Fatalf("expected 1 error message in app log")
	}

	errorLines := readLogLines(t, dir, "error")
	if len(errorLines) != 1 {
		t.Fatalf("expected 1 error log line, got %d", len(errorLines))
	}
	if countMsgOccurrences(errorLines, "error message") != 1 {
		t.Fatalf("expected 1 error message in error log")
	}
	if countMsgOccurrences(errorLines, "warn message") != 0 {
		t.Fatalf("expected 0 warn messages in error log")
	}
}
