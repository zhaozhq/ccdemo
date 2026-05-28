package storage

import (
	"os"
	"testing"
	"time"
)

func tempDB(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	name := f.Name()
	f.Close()
	t.Cleanup(func() { os.Remove(name) })
	return name
}

func TestOpenAndCreateSchema(t *testing.T) {
	dbPath := tempDB(t)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	items, err := s.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestSaveAndList(t *testing.T) {
	dbPath := tempDB(t)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	now := time.Now()
	items := []HotSearch{
		{
			Title:     "Go 1.26 released",
			URL:       "https://example.com/1",
			Platform:  "github",
			Rank:      1,
			Heat:      10000,
			Category:  "tech",
			CreatedAt: now,
		},
		{
			Title:     "SQLite in Go",
			URL:       "https://example.com/2",
			Platform:  "github",
			Rank:      2,
			Heat:      8000,
			Category:  "tech",
			CreatedAt: now,
		},
	}

	if err := s.Save(items); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	all, err := s.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 items, got %d", len(all))
	}
	if all[0].Rank != 1 {
		t.Fatalf("expected first item rank 1, got %d", all[0].Rank)
	}
	if all[1].Rank != 2 {
		t.Fatalf("expected second item rank 2, got %d", all[1].Rank)
	}

	byPlatform, err := s.ListByPlatform("github")
	if err != nil {
		t.Fatalf("ListByPlatform failed: %v", err)
	}
	if len(byPlatform) != 2 {
		t.Fatalf("expected 2 items for github, got %d", len(byPlatform))
	}

	byOther, err := s.ListByPlatform("twitter")
	if err != nil {
		t.Fatalf("ListByPlatform failed: %v", err)
	}
	if len(byOther) != 0 {
		t.Fatalf("expected 0 items for twitter, got %d", len(byOther))
	}
}

func TestDeleteBefore(t *testing.T) {
	dbPath := tempDB(t)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	oldTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cutoff := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	items := []HotSearch{
		{
			Title:     "Old news",
			URL:       "https://example.com/old",
			Platform:  "github",
			Rank:      1,
			Heat:      100,
			Category:  "tech",
			CreatedAt: oldTime,
		},
		{
			Title:     "New news",
			URL:       "https://example.com/new",
			Platform:  "github",
			Rank:      2,
			Heat:      200,
			Category:  "tech",
			CreatedAt: newTime,
		},
	}

	if err := s.Save(items); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := s.DeleteBefore(cutoff); err != nil {
		t.Fatalf("DeleteBefore failed: %v", err)
	}

	remaining, err := s.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 item after delete, got %d", len(remaining))
	}
	if remaining[0].Title != "New news" {
		t.Fatalf("expected remaining item 'New news', got %s", remaining[0].Title)
	}
}
