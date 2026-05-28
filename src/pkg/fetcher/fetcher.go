package fetcher

import (
	"context"
	"time"
)

// Item represents a single fetched hot-search item.
type Item struct {
	Title    string
	URL      string
	Rank     int
	Heat     int64
	Category string
}

// FetchResult is the aggregated result of a fetch operation.
type FetchResult struct {
	Platform  string
	Items     []Item
	FetchedAt time.Time
	Error     error
}

// Fetcher defines the interface for platform-specific fetchers.
type Fetcher interface {
	Fetch(ctx context.Context) (FetchResult, error)
	Name() string
}
