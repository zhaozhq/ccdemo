package mcp

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"ccdemo/src/pkg/storage"

	"github.com/mark3labs/mcp-go/mcp"
)

func tempDB(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "mcp-test-*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	name := f.Name()
	_ = f.Close()
	t.Cleanup(func() { os.Remove(name) })
	return name
}

func sampleStorage(t *testing.T) *storage.Storage {
	t.Helper()
	dbPath := tempDB(t)
	s, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	now := time.Now()
	items := []storage.HotSearch{
		{Title: "Go 1.26 released", URL: "https://example.com/1", Platform: "weibo", Rank: 1, Heat: 10000, Category: "tech", CreatedAt: now},
		{Title: "SQLite in Go", URL: "https://example.com/2", Platform: "baidu", Rank: 2, Heat: 8000, Category: "tech", CreatedAt: now},
		{Title: "Zap logger", URL: "https://example.com/3", Platform: "zhihu", Rank: 1, Heat: 5000, Category: "tech", CreatedAt: now},
	}
	if err := s.Save(items); err != nil {
		t.Fatalf("save items: %v", err)
	}
	return s
}

func TestNewServer(t *testing.T) {
	store := sampleStorage(t)
	srv := NewServer(store)
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
	if srv.store != store {
		t.Error("expected store to be set")
	}
	if srv.server == nil {
		t.Error("expected internal MCPServer to be initialized")
	}
}

func TestGetPlatformsTool(t *testing.T) {
	store := sampleStorage(t)
	srv := NewServer(store)

	// Verify that calling the handler does not panic and returns expected platforms.
	ctx := context.Background()
	req := struct {
		Method string                 `json:"method"`
		Params map[string]interface{} `json:"params"`
	}{
		Method: "tools/call",
		Params: map[string]interface{}{
			"name":      "get_platforms",
			"arguments": map[string]interface{}{},
		},
	}

	// We call the internal handler directly to avoid full MCP protocol negotiation.
	result, err := srv.handleGetPlatforms(ctx, toCallToolRequest(req))
	if err != nil {
		t.Fatalf("handleGetPlatforms error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}

	text := getTextFromFirstContent(t, result)
	var platforms []string
	if err := json.Unmarshal([]byte(text), &platforms); err != nil {
		t.Fatalf("unmarshal platforms: %v", err)
	}
	expected := []string{"weibo", "baidu", "zhihu", "douyin"}
	if len(platforms) != len(expected) {
		t.Fatalf("expected %d platforms, got %d", len(expected), len(platforms))
	}
	for i, v := range expected {
		if platforms[i] != v {
			t.Errorf("expected platform %q at index %d, got %q", v, i, platforms[i])
		}
	}
}

func TestGetHotSearchesTool(t *testing.T) {
	store := sampleStorage(t)
	srv := NewServer(store)

	ctx := context.Background()
	req := struct {
		Method string                 `json:"method"`
		Params map[string]interface{} `json:"params"`
	}{
		Method: "tools/call",
		Params: map[string]interface{}{
			"name":      "get_hot_searches",
			"arguments": map[string]interface{}{},
		},
	}

	result, err := srv.handleGetHotSearches(ctx, toCallToolRequest(req))
	if err != nil {
		t.Fatalf("handleGetHotSearches error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	text := getTextFromFirstContent(t, result)
	var items []storage.HotSearch
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Fatalf("unmarshal hot searches: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 hot searches, got %d", len(items))
	}
}

func toCallToolRequest(req any) mcp.CallToolRequest {
	data, _ := json.Marshal(req)
	var ctr mcp.CallToolRequest
	_ = json.Unmarshal(data, &ctr)
	return ctr
}

func getTextFromFirstContent(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no content in result")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

func TestGetHotSearchesByPlatformTool(t *testing.T) {
	store := sampleStorage(t)
	srv := NewServer(store)

	ctx := context.Background()
	req := struct {
		Method string                 `json:"method"`
		Params map[string]interface{} `json:"params"`
	}{
		Method: "tools/call",
		Params: map[string]interface{}{
			"name": "get_hot_searches_by_platform",
			"arguments": map[string]interface{}{
				"platform": "weibo",
			},
		},
	}

	result, err := srv.handleGetHotSearchesByPlatform(ctx, toCallToolRequest(req))
	if err != nil {
		t.Fatalf("handleGetHotSearchesByPlatform error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	text := getTextFromFirstContent(t, result)
	var items []storage.HotSearch
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Fatalf("unmarshal hot searches: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 hot search for weibo, got %d", len(items))
	}
	if items[0].Platform != "weibo" {
		t.Errorf("expected platform weibo, got %s", items[0].Platform)
	}
}
