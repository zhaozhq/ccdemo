package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// timeNow is a package-level variable for time injection in tests.
var timeNow = time.Now

// DailyRotateWriter writes log data to files that rotate daily.
// It implements zapcore.WriteSyncer (io.Writer + Sync() error).
type DailyRotateWriter struct {
	dir    string
	prefix string
	ext    string

	mu   sync.Mutex
	file *os.File
	date string // current date in YYYY-MM-DD
}

// NewDailyRotateWriter creates a new DailyRotateWriter.
func NewDailyRotateWriter(dir, prefix string) *DailyRotateWriter {
	return &DailyRotateWriter{
		dir:    dir,
		prefix: prefix,
		ext:    ".log",
	}
}

// Write writes p to the current log file, rotating if the date has changed.
func (w *DailyRotateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	today := timeNow().Format("2006-01-02")
	if w.date != today {
		if w.file != nil {
			_ = w.file.Close()
			w.file = nil
		}
		if err := os.MkdirAll(w.dir, 0755); err != nil {
			return 0, fmt.Errorf("create log dir: %w", err)
		}
		fname := filepath.Join(w.dir, w.prefix+"."+today+w.ext)
		f, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return 0, fmt.Errorf("open log file: %w", err)
		}
		w.file = f
		w.date = today
	}

	return w.file.Write(p)
}

// Sync flushes the current file to disk.
func (w *DailyRotateWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// Close closes the current log file.
func (w *DailyRotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		w.date = ""
		return err
	}
	return nil
}
