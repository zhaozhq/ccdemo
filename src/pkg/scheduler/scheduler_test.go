package scheduler

import (
	"context"
	"testing"
	"time"

	"ccdemo/src/pkg/fetcher"
	"ccdemo/src/pkg/storage"
)

type mockFetcher struct {
	name  string
	items []fetcher.Item
	err   error
}

func (m *mockFetcher) Name() string { return m.name }
func (m *mockFetcher) Fetch(ctx context.Context) (fetcher.FetchResult, error) {
	if m.err != nil {
		return fetcher.FetchResult{}, m.err
	}
	return fetcher.FetchResult{Platform: m.name, Items: m.items, FetchedAt: time.Now()}, nil
}

type mockNotifier struct {
	sent   int
	chatID int64
	items  []storage.HotSearch
}

func (m *mockNotifier) SendHotSearch(chatID int64, items []storage.HotSearch) error {
	m.sent++
	m.chatID = chatID
	m.items = items
	return nil
}

func TestFetchAndStore(t *testing.T) {
	store, err := storage.Open(":memory:")
	if err != nil {
		t.Fatalf("open memory db: %v", err)
	}
	defer store.Close()

	f1 := &mockFetcher{
		name: "platA",
		items: []fetcher.Item{
			{Title: "A1", URL: "http://a1", Rank: 1, Heat: 100, Category: "c1"},
			{Title: "A2", URL: "http://a2", Rank: 2, Heat: 90, Category: "c2"},
		},
	}
	f2 := &mockFetcher{
		name: "platB",
		items: []fetcher.Item{
			{Title: "B1", URL: "http://b1", Rank: 1, Heat: 95, Category: "c1"},
		},
	}

	ntf := &mockNotifier{}
	sch := New(store, []fetcher.Fetcher{f1, f2}, ntf, 12345)

	ctx := context.Background()
	if err := sch.FetchAndStore(ctx); err != nil {
		t.Fatalf("FetchAndStore error: %v", err)
	}

	items, err := store.ListAll()
	if err != nil {
		t.Fatalf("ListAll error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items stored, got %d", len(items))
	}
}

func TestPush(t *testing.T) {
	store, err := storage.Open(":memory:")
	if err != nil {
		t.Fatalf("open memory db: %v", err)
	}
	defer store.Close()

	now := time.Now()
	items := []storage.HotSearch{
		{Title: "T1", URL: "http://t1", Platform: "p1", Rank: 1, Heat: 100, Category: "c1", CreatedAt: now},
		{Title: "T2", URL: "http://t2", Platform: "p1", Rank: 2, Heat: 80, Category: "c2", CreatedAt: now},
	}
	if err := store.Save(items); err != nil {
		t.Fatalf("save items: %v", err)
	}

	ntf := &mockNotifier{}
	sch := New(store, nil, ntf, 67890)

	if err := sch.Push(); err != nil {
		t.Fatalf("Push error: %v", err)
	}

	if ntf.sent != 1 {
		t.Fatalf("expected notifier sent=1, got %d", ntf.sent)
	}
	if ntf.chatID != 67890 {
		t.Fatalf("expected chatID 67890, got %d", ntf.chatID)
	}
	if len(ntf.items) != 2 {
		t.Fatalf("expected 2 items sent, got %d", len(ntf.items))
	}
}
