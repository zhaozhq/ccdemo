package fetcher

import (
	"context"
	"time"
)

// DouyinFetcher is a stub fetcher for Douyin (no reliable public API available).
type DouyinFetcher struct{}

// NewDouyinFetcher creates a new DouyinFetcher.
func NewDouyinFetcher() *DouyinFetcher {
	return &DouyinFetcher{}
}

// Name returns the platform name.
func (d *DouyinFetcher) Name() string {
	return "douyin"
}

// Fetch returns an empty FetchResult with no error.
func (d *DouyinFetcher) Fetch(ctx context.Context) (FetchResult, error) {
	return FetchResult{
		Platform:  d.Name(),
		Items:     []Item{},
		FetchedAt: time.Now(),
		Error:     nil,
	}, nil
}
