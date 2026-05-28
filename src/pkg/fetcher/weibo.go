package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	weiboAPI     = "https://weibo.com/ajax/side/hotSearch"
	weiboBaseURL = "https://s.weibo.com/weibo"
	maxItems     = 20
)

// WeiboFetcher fetches hot search data from Weibo.
type WeiboFetcher struct {
	client *http.Client
}

// NewWeiboFetcher creates a new WeiboFetcher with a 10s timeout HTTP client.
func NewWeiboFetcher() *WeiboFetcher {
	return &WeiboFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns the platform name.
func (w *WeiboFetcher) Name() string {
	return "weibo"
}

// weiboResponse mirrors the JSON structure returned by Weibo's hot search API.
type weiboResponse struct {
	Data struct {
		Realtime []struct {
			Word   string `json:"word"`
			RawHot int64  `json:"raw_hot"`
		} `json:"realtime"`
	} `json:"data"`
}

// Fetch retrieves the top hot search items from Weibo.
func (w *WeiboFetcher) Fetch(ctx context.Context) (FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, weiboAPI, nil)
	if err != nil {
		return FetchResult{Platform: w.Name(), FetchedAt: time.Now(), Error: err}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Fetcher/1.0)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://weibo.com/")

	resp, err := w.client.Do(req)
	if err != nil {
		return FetchResult{Platform: w.Name(), FetchedAt: time.Now(), Error: err}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("weibo API returned status %d", resp.StatusCode)
		return FetchResult{Platform: w.Name(), FetchedAt: time.Now(), Error: err}, err
	}

	var parsed weiboResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return FetchResult{Platform: w.Name(), FetchedAt: time.Now(), Error: err}, err
	}

	items := make([]Item, 0, maxItems)
	for i, entry := range parsed.Data.Realtime {
		if i >= maxItems {
			break
		}
		items = append(items, Item{
			Title: entry.Word,
			URL:   weiboBaseURL + "?q=" + url.QueryEscape(entry.Word),
			Rank:  i + 1,
			Heat:  entry.RawHot,
		})
	}

	return FetchResult{
		Platform:  w.Name(),
		Items:     items,
		FetchedAt: time.Now(),
		Error:     nil,
	}, nil
}
