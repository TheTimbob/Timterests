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

		if doc.Find(".about-tabs").Length() == 0 {
			t.Error("expected tab nav to be rendered")
		}
	})

	t.Run("returns bio tab partial for HTMX request", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/about?tab=bio", nil)
		rec := httptest.NewRecorder()

		web.AboutHandler(rec, req, *s)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if !strings.Contains(rec.Body.String(), "about-profile") {
			t.Error("expected bio tab to contain profile card")
		}
	})
}

func TestAboutForm(t *testing.T) {
	t.Parallel()

	t.Run("profile card absent in bio tab when all fields empty", func(t *testing.T) {
		t.Parallel()

		about := web.About{
			Title: "About",
			Body:  "<p>Some body text.</p>",
		}

		var buf bytes.Buffer

		err := web.BioTab(about).Render(context.Background(), &buf)
		if err != nil {
			t.Fatalf("render failed: %v", err)
		}

		if strings.Contains(buf.String(), "about-profile-row") {
			t.Error("expected no profile rows when all fields are empty")
		}
	})

	t.Run("bio tab renders populated profile fields", func(t *testing.T) {
		t.Parallel()

		about := web.About{
			Title:     "About",
			Body:      "<p>Some body text.</p>",
			Name:      "Tim Scott",
			Specialty: "Software Engineering",
			Location:  "United States",
			GitHub:    "TheTimbob",
			Email:     "tscott1275@gmail.com",
		}

		var buf bytes.Buffer

		err := web.BioTab(about).Render(context.Background(), &buf)
		if err != nil {
			t.Fatalf("render failed: %v", err)
		}

		html := buf.String()

		for _, want := range []string{
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

	t.Run("experience tab renders timeline items", func(t *testing.T) {
		t.Parallel()

		jobs := []web.Experience{
			{
				Company:   "Acme Corp",
				Role:      "Software Engineer",
				StartDate: "2020",
				EndDate:   "Present",
				Location:  "Remote",
			},
		}

		var buf bytes.Buffer

		err := web.ExperienceTab(jobs).Render(context.Background(), &buf)
		if err != nil {
			t.Fatalf("render failed: %v", err)
		}

		html := buf.String()

		for _, want := range []string{"Acme Corp", "Software Engineer", "2020", "Remote"} {
			if !strings.Contains(html, want) {
				t.Errorf("expected experience tab to contain %q", want)
			}
		}
	})

	t.Run("skills tab renders sections with tags and description", func(t *testing.T) {
		t.Parallel()

		skills := []web.Skill{
			{
				Name:        "Backend",
				Items:       []string{"Go", "Python", "SQL"},
				Description: "Five years building backend systems.",
			},
		}

		var buf bytes.Buffer

		err := web.SkillsTab(skills).Render(context.Background(), &buf)
		if err != nil {
			t.Fatalf("render failed: %v", err)
		}

		html := buf.String()

		for _, want := range []string{"Backend", "Go", "Python", "SQL", "skill-section-title", "Five years"} {
			if !strings.Contains(html, want) {
				t.Errorf("expected skills tab to contain %q", want)
			}
		}
	})
}
