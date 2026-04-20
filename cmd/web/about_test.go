package web_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"timterests/cmd/web"

	"github.com/PuerkitoBio/goquery"
)

func TestAboutHandler(t *testing.T) {
	s := testSetup(t, context.Background())

	t.Run("renders about page", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/about", nil)
		rec := httptest.NewRecorder()

		web.AboutHandler(rec, req, *s)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if doc.Find("title").Length() == 0 {
			t.Error("expected title element to be rendered")
		}
	})
}

func TestAboutForm(t *testing.T) {
	t.Parallel()

	t.Run("profile card absent when all fields empty", func(t *testing.T) {
		t.Parallel()

		about := web.About{
			Title:    "About",
			Subtitle: "A subtitle",
			Body:     "<p>Some body text.</p>",
		}

		var buf bytes.Buffer

		err := web.AboutForm(about).Render(context.Background(), &buf)
		if err != nil {
			t.Fatalf("render failed: %v", err)
		}

		if strings.Contains(buf.String(), "about-profile") {
			t.Error("expected profile card to be absent when all profile fields are empty")
		}
	})

	t.Run("profile card renders all populated fields", func(t *testing.T) {
		t.Parallel()

		about := web.About{
			Title:     "About",
			Subtitle:  "A subtitle",
			Body:      "<p>Some body text.</p>",
			Name:      "Tim Scott",
			Specialty: "Software Engineering",
			Location:  "United States",
			GitHub:    "TheTimbob",
			Email:     "tscott1275@gmail.com",
		}

		var buf bytes.Buffer

		err := web.AboutForm(about).Render(context.Background(), &buf)
		if err != nil {
			t.Fatalf("render failed: %v", err)
		}

		html := buf.String()

		for _, want := range []string{
			"about-profile",
			"Tim Scott",
			"Software Engineering",
			"United States",
			"TheTimbob",
			"tscott1275@gmail.com",
			"https://github.com/TheTimbob",
			"mailto:tscott1275@gmail.com",
		} {
			if !strings.Contains(html, want) {
				t.Errorf("expected rendered output to contain %q", want)
			}
		}
	})
}
