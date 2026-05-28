package aggregator

import (
	"sort"
	"strings"

	"ccdemo/src/pkg/fetcher"
	"ccdemo/src/pkg/storage"
)

// Aggregator merges fetcher results into a unified hot-search list.
type Aggregator struct{}

// New creates a new Aggregator.
func New() *Aggregator {
	return &Aggregator{}
}

// Aggregate merges fetcher results:
//  1. Deduplicate by Title (case-sensitive, trim whitespace)
//  2. Skip empty titles
//  3. Sort by Heat descending
//  4. Reassign Rank starting from 1
//  5. Return top N items (if topN > 0 and len > topN)
//  6. Convert to []storage.HotSearch, set CreatedAt to result[0].FetchedAt
func (a *Aggregator) Aggregate(results []fetcher.FetchResult, topN int) []storage.HotSearch {
	if len(results) == 0 {
		return []storage.HotSearch{}
	}

	// Determine CreatedAt from the first result's FetchedAt.
	createdAt := results[0].FetchedAt

	// Deduplicate by trimmed title, keeping the first seen (highest heat if
	// platforms are ordered by priority, but here we just keep first occurrence).
	seen := make(map[string]struct{}, 64)
	merged := make([]storage.HotSearch, 0, 64)

	for _, res := range results {
		for _, it := range res.Items {
			title := strings.TrimSpace(it.Title)
			if title == "" {
				continue
			}
			if _, ok := seen[title]; ok {
				continue
			}
			seen[title] = struct{}{}
			merged = append(merged, storage.HotSearch{
				Title:     title,
				URL:       it.URL,
				Platform:  res.Platform,
				Heat:      it.Heat,
				Category:  it.Category,
				CreatedAt: createdAt,
			})
		}
	}

	// Sort by Heat descending.
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Heat > merged[j].Heat
	})

	// Reassign Rank starting from 1.
	for i := range merged {
		merged[i].Rank = i + 1
	}

	// Return top N if requested.
	if topN > 0 && len(merged) > topN {
		merged = merged[:topN]
	}

	return merged
}
