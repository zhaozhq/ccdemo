package fetcher

import (
	"context"
	"testing"
	"time"
)

func TestWeiboFetcherImplementsFetcher(t *testing.T) {
	var _ Fetcher = (*WeiboFetcher)(nil)
}

func TestWeiboFetcherName(t *testing.T) {
	f := NewWeiboFetcher()
	if got := f.Name(); got != "weibo" {
		t.Errorf("Name() = %q, want %q", got, "weibo")
	}
}

func TestFetchResultIsValid(t *testing.T) {
	now := time.Now()
	result := FetchResult{
		Platform:  "weibo",
		Items:     []Item{{Title: "test", URL: "https://example.com", Rank: 1, Heat: 100, Category: "hot"}},
		FetchedAt: now,
		Error:     nil,
	}

	if result.Platform != "weibo" {
		t.Errorf("Platform = %q, want %q", result.Platform, "weibo")
	}
	if len(result.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(result.Items))
	}
	if result.Items[0].Title != "test" {
		t.Errorf("Items[0].Title = %q, want %q", result.Items[0].Title, "test")
	}
	if result.Items[0].URL != "https://example.com" {
		t.Errorf("Items[0].URL = %q, want %q", result.Items[0].URL, "https://example.com")
	}
	if result.Items[0].Rank != 1 {
		t.Errorf("Items[0].Rank = %d, want 1", result.Items[0].Rank)
	}
	if result.Items[0].Heat != 100 {
		t.Errorf("Items[0].Heat = %d, want 100", result.Items[0].Heat)
	}
	if result.Items[0].Category != "hot" {
		t.Errorf("Items[0].Category = %q, want %q", result.Items[0].Category, "hot")
	}
	if !result.FetchedAt.Equal(now) {
		t.Errorf("FetchedAt mismatch")
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
}

func TestWeiboFetcherFetchTimeout(t *testing.T) {
	f := NewWeiboFetcher()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := f.Fetch(ctx)
	if err == nil {
		t.Error("expected error for timeout context, got nil")
	}
}

func TestBaiduFetcherImplementsFetcher(t *testing.T) {
	var _ Fetcher = (*BaiduFetcher)(nil)
}

func TestZhihuFetcherImplementsFetcher(t *testing.T) {
	var _ Fetcher = (*ZhihuFetcher)(nil)
}

func TestDouyinFetcherImplementsFetcher(t *testing.T) {
	var _ Fetcher = (*DouyinFetcher)(nil)
}
