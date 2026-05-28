package aggregator

import (
	"testing"
	"time"

	"ccdemo/src/pkg/fetcher"
)

func TestAggregate(t *testing.T) {
	now := time.Now()
	results := []fetcher.FetchResult{
		{
			Platform: "weibo",
			Items: []fetcher.Item{
				{Title: "A", Rank: 1, Heat: 100},
				{Title: "B", Rank: 2, Heat: 90},
			},
			FetchedAt: now,
		},
		{
			Platform: "baidu",
			Items: []fetcher.Item{
				{Title: "B", Rank: 1, Heat: 95}, // duplicate "B"
				{Title: "C", Rank: 2, Heat: 80},
			},
			FetchedAt: now,
		},
	}

	a := New()
	got := a.Aggregate(results, 20)

	if len(got) != 3 {
		t.Fatalf("expected 3 items, got %d", len(got))
	}

	// Should be sorted by Heat descending: A(100), B(90 from weibo), C(80)
	expected := []struct {
		Title    string
		Platform string
		Heat     int64
		Rank     int
	}{
		{"A", "weibo", 100, 1},
		{"B", "weibo", 90, 2},
		{"C", "baidu", 80, 3},
	}

	for i, exp := range expected {
		if got[i].Title != exp.Title {
			t.Errorf("item[%d].Title = %q, want %q", i, got[i].Title, exp.Title)
		}
		if got[i].Platform != exp.Platform {
			t.Errorf("item[%d].Platform = %q, want %q", i, got[i].Platform, exp.Platform)
		}
		if got[i].Heat != exp.Heat {
			t.Errorf("item[%d].Heat = %d, want %d", i, got[i].Heat, exp.Heat)
		}
		if got[i].Rank != exp.Rank {
			t.Errorf("item[%d].Rank = %d, want %d", i, got[i].Rank, exp.Rank)
		}
		if !got[i].CreatedAt.Equal(now) {
			t.Errorf("item[%d].CreatedAt = %v, want %v", i, got[i].CreatedAt, now)
		}
	}
}

func TestAggregateTopN(t *testing.T) {
	now := time.Now()
	results := []fetcher.FetchResult{
		{
			Platform: "weibo",
			Items: []fetcher.Item{
				{Title: "A", Rank: 1, Heat: 100},
				{Title: "B", Rank: 2, Heat: 90},
				{Title: "C", Rank: 3, Heat: 80},
			},
			FetchedAt: now,
		},
	}

	a := New()
	got := a.Aggregate(results, 2)

	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0].Title != "A" || got[1].Title != "B" {
		t.Errorf("expected top 2 [A, B], got [%s, %s]", got[0].Title, got[1].Title)
	}
	if got[0].Rank != 1 || got[1].Rank != 2 {
		t.Errorf("expected ranks [1, 2], got [%d, %d]", got[0].Rank, got[1].Rank)
	}
}

func TestAggregateEmptyTitle(t *testing.T) {
	now := time.Now()
	results := []fetcher.FetchResult{
		{
			Platform: "weibo",
			Items: []fetcher.Item{
				{Title: "A", Rank: 1, Heat: 100},
				{Title: "", Rank: 2, Heat: 90},
				{Title: "  ", Rank: 3, Heat: 80},
			},
			FetchedAt: now,
		},
	}

	a := New()
	got := a.Aggregate(results, 10)

	if len(got) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got))
	}
	if got[0].Title != "A" {
		t.Errorf("expected [A], got [%s]", got[0].Title)
	}
}

func TestAggregateNoResults(t *testing.T) {
	a := New()
	got := a.Aggregate(nil, 10)
	if got == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 items, got %d", len(got))
	}

	got = a.Aggregate([]fetcher.FetchResult{}, 10)
	if got == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 items, got %d", len(got))
	}
}

func TestAggregateCreatedAtFromFirstResult(t *testing.T) {
	now := time.Now()
	results := []fetcher.FetchResult{
		{
			Platform: "weibo",
			Items: []fetcher.Item{
				{Title: "A", Rank: 1, Heat: 100},
			},
			FetchedAt: now,
		},
		{
			Platform: "baidu",
			Items: []fetcher.Item{
				{Title: "B", Rank: 1, Heat: 95},
			},
			FetchedAt: now.Add(time.Hour),
		},
	}

	a := New()
	got := a.Aggregate(results, 10)

	for i, item := range got {
		if !item.CreatedAt.Equal(now) {
			t.Errorf("item[%d].CreatedAt = %v, want %v", i, item.CreatedAt, now)
		}
	}
}
