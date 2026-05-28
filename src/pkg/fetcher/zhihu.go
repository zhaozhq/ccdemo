package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	zhihuAPI = "https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total"
)

// ZhihuFetcher fetches hot search data from Zhihu.
type ZhihuFetcher struct {
	client *http.Client
}

// NewZhihuFetcher creates a new ZhihuFetcher with a 10s timeout HTTP client.
func NewZhihuFetcher() *ZhihuFetcher {
	return &ZhihuFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns the platform name.
func (z *ZhihuFetcher) Name() string {
	return "zhihu"
}

// zhihuResponse mirrors the JSON structure returned by Zhihu's hot list API.
type zhihuResponse struct {
	Data []struct {
		Target struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		} `json:"target"`
		DetailText string `json:"detail_text"`
	} `json:"data"`
}

// Fetch retrieves the top hot search items from Zhihu.
func (z *ZhihuFetcher) Fetch(ctx context.Context) (FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, zhihuAPI, nil)
	if err != nil {
		return FetchResult{Platform: z.Name(), FetchedAt: time.Now(), Error: err}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Fetcher/1.0)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.zhihu.com/")

	resp, err := z.client.Do(req)
	if err != nil {
		return FetchResult{Platform: z.Name(), FetchedAt: time.Now(), Error: err}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("zhihu API returned status %d", resp.StatusCode)
		return FetchResult{Platform: z.Name(), FetchedAt: time.Now(), Error: err}, err
	}

	var parsed zhihuResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return FetchResult{Platform: z.Name(), FetchedAt: time.Now(), Error: err}, err
	}

	items := make([]Item, 0, maxItems)
	for i, entry := range parsed.Data {
		if i >= maxItems {
			break
		}
		items = append(items, Item{
			Title: entry.Target.Title,
			URL:   entry.Target.URL,
			Rank:  i + 1,
		})
	}

	return FetchResult{
		Platform:  z.Name(),
		Items:     items,
		FetchedAt: time.Now(),
		Error:     nil,
	}, nil
}
