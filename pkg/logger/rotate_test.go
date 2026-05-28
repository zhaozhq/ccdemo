package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDailyRotateWriter(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewDailyRotateWriter(tmpDir, "app")
	defer w.Close()

	msg := []byte("hello world\n")
	n, err := w.Write(msg)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(msg) {
		t.Fatalf("Write returned %d, want %d", n, len(msg))
	}

	today := time.Now().Format("2006-01-02")
	expectedName := filepath.Join(tmpDir, "app."+today+".log")
	if _, err := os.Stat(expectedName); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist", expectedName)
	}

	content, err := os.ReadFile(expectedName)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(content) != string(msg) {
		t.Fatalf("file content = %q, want %q", content, msg)
	}
}

func TestDailyRotateWriterCrossDay(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewDailyRotateWriter(tmpDir, "app")
	defer w.Close()

	// Write on "today"
	msg1 := []byte("day one\n")
	if _, err := w.Write(msg1); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Mock time to tomorrow
	tomorrow := time.Now().Add(24 * time.Hour)
	oldTimeNow := timeNow
	timeNow = func() time.Time { return tomorrow }
	defer func() { timeNow = oldTimeNow }()

	msg2 := []byte("day two\n")
	if _, err := w.Write(msg2); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	today := time.Now().Format("2006-01-02")
	tomorrowStr := tomorrow.Format("2006-01-02")

	oldFile := filepath.Join(tmpDir, "app."+today+".log")
	newFile := filepath.Join(tmpDir, "app."+tomorrowStr+".log")

	if _, err := os.Stat(oldFile); os.IsNotExist(err) {
		t.Fatalf("expected old file %s to exist", oldFile)
	}
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Fatalf("expected new file %s to exist", newFile)
	}

	oldContent, _ := os.ReadFile(oldFile)
	if string(oldContent) != string(msg1) {
		t.Fatalf("old file content = %q, want %q", oldContent, msg1)
	}

	newContent, _ := os.ReadFile(newFile)
	if string(newContent) != string(msg2) {
		t.Fatalf("new file content = %q, want %q", newContent, msg2)
	}
}
