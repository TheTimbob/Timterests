package web_test

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timterests/cmd/web"
)

func TestRobotsHandler(t *testing.T) {
	t.Setenv("SITE_URL", "https://example.com")

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/robots.txt", nil)
	rec := httptest.NewRecorder()

	web.RobotsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	for _, want := range []string{
		"User-agent: *",
		"Disallow: /admin",
		"Disallow: /writer",
		"Disallow: /login",
		"Disallow: /download",
		"Sitemap: https://example.com/sitemap.xml",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("robots.txt missing %q", want)
		}
	}
}

func TestSitemapHandler(t *testing.T) {
	s := testSetup(t, context.Background())
	t.Setenv("SITE_URL", "https://example.com")

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/sitemap.xml", nil)
	rec := httptest.NewRecorder()

	web.SitemapHandler(rec, req, *s)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/xml") {
		t.Errorf("expected XML content type, got %q", ct)
	}

	var result struct {
		XMLName xml.Name `xml:"urlset"`
		URLs    []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}

	if err := xml.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse sitemap XML: %v", err)
	}

	if len(result.URLs) < 5 {
		t.Errorf("expected at least 5 URLs in sitemap (static pages), got %d", len(result.URLs))
	}

	hasRoot := false
	for _, u := range result.URLs {
		if u.Loc == "https://example.com/" {
			hasRoot = true
		}
	}

	if !hasRoot {
		t.Error("sitemap missing root URL")
	}
}
