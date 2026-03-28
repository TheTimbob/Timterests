package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"timterests/cmd/web"
	"timterests/internal/auth"
)

// TestIsHTMXRequest verifies that HTMX request detection is accurate.
func TestIsHTMXRequest(t *testing.T) {
	t.Parallel()

	t.Run("returns true when HX-Request header is present", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)
		req.Header.Set("HX-Request", "true")

		if !web.IsHTMXRequest(req) {
			t.Error("expected IsHTMXRequest to return true, got false")
		}
	})

	t.Run("returns false when HX-Request header is absent (back button / direct nav)", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)

		if web.IsHTMXRequest(req) {
			t.Error("expected IsHTMXRequest to return false, got true")
		}
	})

	t.Run("returns false for non-true HX-Request value", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)
		req.Header.Set("HX-Request", "false")

		if web.IsHTMXRequest(req) {
			t.Error("expected IsHTMXRequest to return false for 'false' value, got true")
		}
	})
}

// TestIsBackButtonNavigation verifies back-button detection logic.
func TestIsBackButtonNavigation(t *testing.T) {
	t.Parallel()

	t.Run("returns true when no HX-Request header (back button / direct nav)", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/article?id=1", nil)

		if !web.IsBackButtonNavigation(req) {
			t.Error("expected IsBackButtonNavigation to return true for non-HTMX request, got false")
		}
	})

	t.Run("returns false when HX-Request header is present (normal HTMX swap)", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/article?id=1", nil)
		req.Header.Set("HX-Request", "true")

		if web.IsBackButtonNavigation(req) {
			t.Error("expected IsBackButtonNavigation to return false for HTMX request, got true")
		}
	})
}

// TestSetPartialResponseHeaders verifies that partial responses get cache-prevention headers.
func TestSetPartialResponseHeaders(t *testing.T) {
	t.Parallel()

	t.Run("sets Cache-Control no-store header", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		web.SetPartialResponseHeaders(rec)

		cacheControl := rec.Header().Get("Cache-Control")
		if cacheControl == "" {
			t.Error("expected Cache-Control header to be set, but it was empty")
		}

		// Must contain no-store to prevent back-button from serving cached partial
		if cacheControl != "no-store, no-cache, must-revalidate" {
			t.Errorf("expected Cache-Control 'no-store, no-cache, must-revalidate', got %q", cacheControl)
		}
	})

	t.Run("sets Vary header for HX-Request", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		web.SetPartialResponseHeaders(rec)

		vary := rec.Header().Get("Vary")
		if vary != "HX-Request" {
			t.Errorf("expected Vary 'HX-Request', got %q", vary)
		}
	})
}

// TestArticlesBackButtonBehavior tests that back-button navigation returns a full page.
func TestArticlesBackButtonBehavior(t *testing.T) {
	s := testSetup(t, context.Background())

	t.Run("forward nav (HTMX swap) - returns partial, sets cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		web.ArticlesPageHandler(rec, req, *s, "all", "list")

		// Partial response should NOT contain full page structure
		body := rec.Body.String()
		if contains(body, "<title>") {
			t.Error("HTMX partial should not contain <title> tag")
		}
		if contains(body, "<html") {
			t.Error("HTMX partial should not contain <html> tag")
		}

		// Must have cache-prevention headers
		if rec.Header().Get("Cache-Control") == "" {
			t.Error("partial response must have Cache-Control header to prevent back-button cache")
		}
	})

	t.Run("back button / direct nav - returns full page, no cache override needed", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)
		// No HX-Request header = back button or direct navigation
		rec := httptest.NewRecorder()

		web.ArticlesPageHandler(rec, req, *s, "all", "list")

		// Full page response must include base layout
		body := rec.Body.String()
		if !contains(body, "<title>") {
			t.Error("back-button / direct nav must return full page with <title>")
		}
		if !contains(body, "<html") {
			t.Error("back-button / direct nav must return full page with <html>")
		}
		if !contains(body, "nav-header") {
			t.Error("back-button / direct nav must return full page with navigation")
		}
	})
}

// TestArticleDetailBackButtonBehavior tests the article detail route.
func TestArticleDetailBackButtonBehavior(t *testing.T) {
	s := testSetup(t, context.Background())
	a := auth.NewAuth("test-session-key-minimum-32-bytes")

	t.Run("forward nav (HTMX swap) to article detail - partial + cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/article?id=0", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		web.GetArticleHandler(rec, req, *s, "0", a)

		body := rec.Body.String()
		if contains(body, "<title>") {
			t.Error("HTMX partial should not contain <title>")
		}

		if rec.Header().Get("Cache-Control") == "" {
			t.Error("HTMX partial response must have Cache-Control header")
		}
	})

	t.Run("back button to article detail - full page returned", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/article?id=0", nil)
		// No HX-Request = back button or bookmark
		rec := httptest.NewRecorder()

		web.GetArticleHandler(rec, req, *s, "0", a)

		body := rec.Body.String()
		if !contains(body, "<title>") {
			t.Error("back-button to article must return full page with <title>")
		}
		if !contains(body, "nav-header") {
			t.Error("back-button to article must return full page with navigation")
		}
	})
}

// TestProjectsBackButtonBehavior tests the projects routes for back-button safety.
func TestProjectsBackButtonBehavior(t *testing.T) {
	s := testSetup(t, context.Background())
	a := auth.NewAuth("test-session-key-minimum-32-bytes")

	t.Run("projects list - HTMX partial has cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		web.ProjectsPageHandler(rec, req, *s, "all", "list")

		if rec.Header().Get("Cache-Control") == "" {
			t.Error("HTMX partial for projects must have Cache-Control header")
		}
	})

	t.Run("projects list - back button returns full page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		rec := httptest.NewRecorder()

		web.ProjectsPageHandler(rec, req, *s, "all", "list")

		body := rec.Body.String()
		if !contains(body, "<title>") {
			t.Error("back-button to /projects must return full page")
		}
	})

	t.Run("project detail - HTMX partial has cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/project?id=0", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		web.GetProjectHandler(rec, req, *s, "0", a)

		if rec.Header().Get("Cache-Control") == "" {
			t.Error("HTMX partial for project detail must have Cache-Control header")
		}
	})

	t.Run("project detail - back button returns full page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/project?id=0", nil)
		rec := httptest.NewRecorder()

		web.GetProjectHandler(rec, req, *s, "0", a)

		body := rec.Body.String()
		if !contains(body, "<title>") {
			t.Error("back-button to /project must return full page")
		}
	})
}

// TestReadingListBackButtonBehavior tests reading list routes.
func TestReadingListBackButtonBehavior(t *testing.T) {
	s := testSetup(t, context.Background())
	a := auth.NewAuth("test-session-key-minimum-32-bytes")

	t.Run("reading list - HTMX partial has cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/reading-list", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		web.ReadingListPageHandler(rec, req, *s, "all", "list")

		if rec.Header().Get("Cache-Control") == "" {
			t.Error("HTMX partial for reading list must have Cache-Control header")
		}
	})

	t.Run("reading list - back button returns full page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/reading-list", nil)
		rec := httptest.NewRecorder()

		web.ReadingListPageHandler(rec, req, *s, "all", "list")

		body := rec.Body.String()
		if !contains(body, "<title>") {
			t.Error("back-button to /reading-list must return full page")
		}
	})

	t.Run("book detail - HTMX partial has cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/book?id=0", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		web.GetReadingListBook(rec, req, *s, "0", a)

		if rec.Header().Get("Cache-Control") == "" {
			t.Error("HTMX partial for book detail must have Cache-Control header")
		}
	})

	t.Run("book detail - back button returns full page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/book?id=0", nil)
		rec := httptest.NewRecorder()

		web.GetReadingListBook(rec, req, *s, "0", a)

		body := rec.Body.String()
		if !contains(body, "<title>") {
			t.Error("back-button to /book must return full page")
		}
	})
}

// TestLettersBackButtonBehavior tests letters routes (requires authentication).
func TestLettersBackButtonBehavior(t *testing.T) {
	s := testSetup(t, context.Background())
	a, addAuthCookie := testAuthentication(t)

	t.Run("letters list - HTMX partial has cache headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/letters", nil)
		addAuthCookie(req)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		web.LettersPageHandler(rec, req, *s, "all", "list", a)

		if rec.Header().Get("Cache-Control") == "" {
			t.Error("HTMX partial for letters must have Cache-Control header")
		}
	})

	t.Run("letters list - back button returns full page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/letters", nil)
		addAuthCookie(req)
		rec := httptest.NewRecorder()

		web.LettersPageHandler(rec, req, *s, "all", "list", a)

		body := rec.Body.String()
		if !contains(body, "<title>") {
			t.Error("back-button to /letters must return full page")
		}
	})

	t.Run("letter detail - HTMX partial has cache headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/letter?id=0", nil)
		addAuthCookie(req)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		web.GetLetterHandler(rec, req, *s, "0", a)

		if rec.Header().Get("Cache-Control") == "" {
			t.Error("HTMX partial for letter detail must have Cache-Control header")
		}
	})

	t.Run("letter detail - back button returns full page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/letter?id=0", nil)
		addAuthCookie(req)
		rec := httptest.NewRecorder()

		web.GetLetterHandler(rec, req, *s, "0", a)

		body := rec.Body.String()
		if !contains(body, "<title>") {
			t.Error("back-button to /letter must return full page")
		}
	})
}

// TestServerSideFallbackWithoutCacheHeaders verifies that even if Cache-Control is not set,
// the server still returns full pages for back-button requests (no HX-Request header).
// This is a belt-and-suspenders test: the primary protection is Cache-Control headers,
// but the server-side logic (checking for HX-Request) is the fallback guarantee.
func TestServerSideFallbackWithoutCacheHeaders(t *testing.T) {
	s := testSetup(t, context.Background())
	a := auth.NewAuth("test-session-key-minimum-32-bytes")

	t.Run("articles list - back button returns full page even without cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)
		// No HX-Request header (back button scenario)
		rec := httptest.NewRecorder()

		web.ArticlesPageHandler(rec, req, *s, "all", "list")

		// Even if Cache-Control were missing, the handler logic ensures full page is returned
		body := rec.Body.String()
		if !contains(body, "<html") {
			t.Error("back-button to /articles must return full HTML document (fallback guarantee)")
		}
		if !contains(body, "<title>") {
			t.Error("back-button to /articles must return <title> (fallback guarantee)")
		}
		if !contains(body, "nav-header") {
			t.Error("back-button to /articles must return navigation (fallback guarantee)")
		}
	})

	t.Run("article detail - back button returns full page even without cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/article?id=0", nil)
		// No HX-Request header = back button or bookmark
		rec := httptest.NewRecorder()

		web.GetArticleHandler(rec, req, *s, "0", a)

		body := rec.Body.String()
		if !contains(body, "<html") {
			t.Error("back-button to article must return full HTML (server-side fallback)")
		}
		if !contains(body, "<title>") {
			t.Error("back-button to article must return <title> (server-side fallback)")
		}
	})

	t.Run("projects list - back button returns full page (server-side fallback)", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		rec := httptest.NewRecorder()

		web.ProjectsPageHandler(rec, req, *s, "all", "list")

		body := rec.Body.String()
		if !contains(body, "<html") {
			t.Error("back-button to /projects must return full HTML")
		}
		if !contains(body, "nav-header") {
			t.Error("back-button to /projects must include navigation")
		}
	})

	t.Run("reading list - back button returns full page (server-side fallback)", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/reading-list", nil)
		rec := httptest.NewRecorder()

		web.ReadingListPageHandler(rec, req, *s, "all", "list")

		body := rec.Body.String()
		if !contains(body, "<html") {
			t.Error("back-button to /reading-list must return full HTML")
		}
		if !contains(body, "nav-header") {
			t.Error("back-button to /reading-list must include navigation")
		}
	})
}

// contains is a helper to check if a string contains a substring (avoids import of strings in test).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}

			return false
		}())
}
