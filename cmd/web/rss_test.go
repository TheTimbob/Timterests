package web_test

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"timterests/cmd/web"
)

func TestRSSHandler(t *testing.T) {
	s := testSetup(t, context.Background())
	t.Setenv("SITE_URL", "https://example.com")
	t.Setenv("SITE_NAME", "TestBlog")

	req := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet, "/rss.xml", nil,
	)
	rec := httptest.NewRecorder()

	web.RSSHandler(rec, req, *s)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/rss+xml") {
		t.Errorf("expected rss+xml content type, got %q", ct)
	}

	var result struct {
		XMLName xml.Name `xml:"rss"`
		Version string   `xml:"version,attr"`
		Channel struct {
			Title string `xml:"title"`
			Link  string `xml:"link"`
			Items []struct {
				Title   string `xml:"title"`
				Link    string `xml:"link"`
				GUID    string `xml:"guid"`
				PubDate string `xml:"pubDate"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	err := xml.Unmarshal(rec.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("failed to parse RSS XML: %v", err)
	}

	if result.Version != "2.0" {
		t.Errorf("expected RSS version 2.0, got %q", result.Version)
	}

	if result.Channel.Title != "TestBlog" {
		t.Errorf("expected channel title %q, got %q",
			"TestBlog", result.Channel.Title)
	}

	if result.Channel.Link != "https://example.com" {
		t.Errorf("expected channel link %q, got %q",
			"https://example.com", result.Channel.Link)
	}

	if len(result.Channel.Items) == 0 {
		t.Fatal("expected at least one item, got none")
	}

	for _, item := range result.Channel.Items {
		if item.Title == "" {
			t.Error("RSS item has empty title")
		}

		if !strings.HasPrefix(item.Link, "https://example.com/") {
			t.Errorf("RSS item link missing base URL: %q", item.Link)
		}

		_, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			t.Errorf("RSS item pubDate %q is not valid RFC 822: %v", item.PubDate, err)
		}
	}
}
