package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	baiduAPI = "https://top.baidu.com/api/board?platform=wise&tab=realtime"
)

// BaiduFetcher fetches hot search data from Baidu.
type BaiduFetcher struct {
	client *http.Client
}

// NewBaiduFetcher creates a new BaiduFetcher with a 10s timeout HTTP client.
func NewBaiduFetcher() *BaiduFetcher {
	return &BaiduFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns the platform name.
func (b *BaiduFetcher) Name() string {
	return "baidu"
}

// baiduResponse mirrors the JSON structure returned by Baidu's hot search API.
type baiduResponse struct {
	Data struct {
		Cards []struct {
			Content []struct {
				Word     string `json:"word"`
				RawURL   string `json:"rawUrl"`
				HotScore int64  `json:"hotScore"`
			} `json:"content"`
		} `json:"cards"`
	} `json:"data"`
}

// Fetch retrieves the top hot search items from Baidu.
func (b *BaiduFetcher) Fetch(ctx context.Context) (FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baiduAPI, nil)
	if err != nil {
		return FetchResult{Platform: b.Name(), FetchedAt: time.Now(), Error: err}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Fetcher/1.0)")
	req.Header.Set("Accept", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return FetchResult{Platform: b.Name(), FetchedAt: time.Now(), Error: err}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("baidu API returned status %d", resp.StatusCode)
		return FetchResult{Platform: b.Name(), FetchedAt: time.Now(), Error: err}, err
	}

	var parsed baiduResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return FetchResult{Platform: b.Name(), FetchedAt: time.Now(), Error: err}, err
	}

	items := make([]Item, 0, maxItems)
	for _, card := range parsed.Data.Cards {
		for i, entry := range card.Content {
			if i >= maxItems {
				break
			}
			items = append(items, Item{
				Title: entry.Word,
				URL:   entry.RawURL,
				Rank:  i + 1,
				Heat:  entry.HotScore,
			})
		}
		if len(items) >= maxItems {
			break
		}
	}

	return FetchResult{
		Platform:  b.Name(),
		Items:     items,
		FetchedAt: time.Now(),
		Error:     nil,
	}, nil
}
