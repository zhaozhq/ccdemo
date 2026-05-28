package bot

import (
	"testing"

	"ccdemo/src/pkg/storage"
)

func TestFormatMessage(t *testing.T) {
	items := []storage.HotSearch{
		{
			Title:    "Test Title 1",
			URL:      "https://example.com/1",
			Platform: "weibo",
			Heat:     1000000,
		},
		{
			Title:    "Test Title 2",
			URL:      "https://example.com/2",
			Platform: "zhihu",
			Heat:     500000,
		},
	}

	got := formatMessage(items)

	want := "*今日热搜 Top20*\n\n" +
		"1. [Test Title 1](https://example.com/1) 🔥1000000 (weibo)\n" +
		"2. [Test Title 2](https://example.com/2) 🔥500000 (zhihu)\n"

	if got != want {
		t.Errorf("formatMessage() = %q, want %q", got, want)
	}
}

func TestFormatMessageEmpty(t *testing.T) {
	items := []storage.HotSearch{}

	got := formatMessage(items)

	want := "暂无热搜数据。"

	if got != want {
		t.Errorf("formatMessage() = %q, want %q", got, want)
	}
}
