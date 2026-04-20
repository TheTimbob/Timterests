package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"timterests/cmd/web"
	"timterests/internal/auth"
)

// TestIsHTMXRequest verifies that HTMX request detection is accurate.
func TestIsHTMXRequest(t *testing.T) {
	t.Parallel()

	t.Run("returns true when HX-Request header is present", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/articles", nil)
		req.Header.Set("Hx-Request", "true")

		if !web.IsHTMXRequest(req) {
			t.Error("expected IsHTMXRequest to return true, got false")
		}
	})

	t.Run("returns false when HX-Request header is absent (back button / direct nav)", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/articles", nil)

		if web.IsHTMXRequest(req) {
			t.Error("expected IsHTMXRequest to return false, got true")
		}
	})

	t.Run("returns false for non-true HX-Request value", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/articles", nil)
		req.Header.Set("Hx-Request", "false")

		if web.IsHTMXRequest(req) {
			t.Error("expected IsHTMXRequest to return false for 'false' value, got true")
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
		if !strings.Contains(cacheControl, "no-store") {
			t.Errorf("expected Cache-Control to contain 'no-store', got %q", cacheControl)
		}
	})

	t.Run("sets Vary header for HX-Request", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		web.SetPartialResponseHeaders(rec)

		vary := rec.Header().Get("Vary")
		if !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain 'HX-Request', got %q", vary)
		}
	})

	t.Run("does not append to Vary: * (RFC 7231 terminal value)", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		rec.Header().Set("Vary", "*")
		web.SetPartialResponseHeaders(rec)

		vary := rec.Header().Get("Vary")
		if vary != "*" {
			t.Errorf("expected Vary to remain '*' when pre-set, got %q", vary)
		}
	})

	t.Run("does not duplicate HX-Request in Vary when already present", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		rec.Header().Set("Vary", "Accept-Encoding, HX-Request")
		web.SetPartialResponseHeaders(rec)

		vary := rec.Header().Get("Vary")
		count := strings.Count(vary, "HX-Request")

		if count != 1 {
			t.Errorf("expected exactly one HX-Request in Vary, got %d in %q", count, vary)
		}
	})
}

func TestFallbackFullPageBehavior(t *testing.T) {
	s := testSetup(t, context.Background())
	a, addAuthCookie := testAuthentication(t)

	routes := []struct {
		name    string
		handler func(rec *httptest.ResponseRecorder, req *http.Request)
		path    string
	}{
		{
			name: "articles list",
			path: "/articles",
			handler: func(rec *httptest.ResponseRecorder, req *http.Request) {
				web.ArticlesPageHandler(rec, req, *s, "all", "list")
			},
		},
		{
			name: "article detail",
			path: "/article?id=0",
			handler: func(rec *httptest.ResponseRecorder, req *http.Request) {
				web.GetArticleHandler(rec, req, *s, "0", a)
			},
		},
		{
			name: "projects list",
			path: "/projects",
			handler: func(rec *httptest.ResponseRecorder, req *http.Request) {
				web.ProjectsPageHandler(rec, req, *s, "all", "list")
			},
		},
		{
			name: "project detail",
			path: "/project?id=0",
			handler: func(rec *httptest.ResponseRecorder, req *http.Request) {
				web.GetProjectHandler(rec, req, *s, "0", a)
			},
		},
		{
			name: "reading list",
			path: "/reading-list",
			handler: func(rec *httptest.ResponseRecorder, req *http.Request) {
				web.ReadingListPageHandler(rec, req, *s, "all", "list")
			},
		},
		{
			name: "book detail",
			path: "/book?id=0",
			handler: func(rec *httptest.ResponseRecorder, req *http.Request) {
				web.GetReadingListBook(rec, req, *s, "0", a)
			},
		},
	}

	for _, route := range routes {
		route := route
		t.Run(route.name+" - no HX-Request returns full page", func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, route.path, nil)
			addAuthCookie(req)
			// Deliberately omit HX-Request header to simulate back-button / direct nav
			rec := httptest.NewRecorder()

			route.handler(rec, req)

			body := rec.Body.String()
			if !strings.Contains(body, "<title>") {
				t.Errorf("%s: expected full page with <title> when HX-Request is absent, got partial", route.name)
			}

			if !strings.Contains(body, "<html") {
				t.Errorf("%s: expected full page with <html> when HX-Request is absent, got partial", route.name)
			}
		})
	}
}

// TestArticlesBackButtonBehavior tests that back-button navigation returns a full page.
func TestArticlesBackButtonBehavior(t *testing.T) {
	s := testSetup(t, context.Background())

	t.Run("forward nav (HTMX swap) - returns partial, sets cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/articles", nil)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.ArticlesPageHandler(rec, req, *s, "all", "list")

		// Partial response should NOT contain full page structure
		body := rec.Body.String()
		if strings.Contains(body, "<title>") {
			t.Error("HTMX partial should not contain <title> tag")
		}

		if strings.Contains(body, "<html") {
			t.Error("HTMX partial should not contain <html> tag")
		}

		// Must have cache-prevention headers
		cacheControl := rec.Header().Get("Cache-Control")
		if cacheControl == "" {
			t.Error("partial response must have Cache-Control header to prevent back-button cache")
		}

		if !strings.Contains(cacheControl, "no-store") {
			t.Errorf("expected Cache-Control to contain 'no-store', got %q", cacheControl)
		}

		vary := rec.Header().Get("Vary")
		if !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain 'HX-Request', got %q", vary)
		}
	})

	t.Run("back button / direct nav - returns full page, no cache override needed", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/articles", nil)
		// No HX-Request header = back button or direct navigation
		rec := httptest.NewRecorder()

		web.ArticlesPageHandler(rec, req, *s, "all", "list")

		// Full page response must include base layout
		body := rec.Body.String()
		if !strings.Contains(body, "<title>") {
			t.Error("back-button / direct nav must return full page with <title>")
		}

		if !strings.Contains(body, "<html") {
			t.Error("back-button / direct nav must return full page with <html>")
		}

		if !strings.Contains(body, "nav-header") {
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

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/article?id=0", nil)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.GetArticleHandler(rec, req, *s, "0", a)

		body := rec.Body.String()
		if strings.Contains(body, "<title>") {
			t.Error("HTMX partial should not contain <title>")
		}

		cacheControl := rec.Header().Get("Cache-Control")
		if cacheControl == "" {
			t.Error("HTMX partial response must have Cache-Control header")
		}

		if !strings.Contains(cacheControl, "no-store") {
			t.Errorf("expected Cache-Control to contain 'no-store', got %q", cacheControl)
		}

		vary := rec.Header().Get("Vary")
		if !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain 'HX-Request', got %q", vary)
		}
	})

	t.Run("back button to article detail - full page returned", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/article?id=0", nil)
		// No HX-Request = back button or bookmark
		rec := httptest.NewRecorder()

		web.GetArticleHandler(rec, req, *s, "0", a)

		body := rec.Body.String()
		if !strings.Contains(body, "<title>") {
			t.Error("back-button to article must return full page with <title>")
		}

		if !strings.Contains(body, "nav-header") {
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

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/projects", nil)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.ProjectsPageHandler(rec, req, *s, "all", "list")

		cacheControl := rec.Header().Get("Cache-Control")
		if cacheControl == "" {
			t.Error("HTMX partial for projects must have Cache-Control header")
		}

		if !strings.Contains(cacheControl, "no-store") {
			t.Errorf("expected Cache-Control to contain 'no-store', got %q", cacheControl)
		}

		vary := rec.Header().Get("Vary")
		if !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain 'HX-Request', got %q", vary)
		}
	})

	t.Run("projects list - back button returns full page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/projects", nil)
		rec := httptest.NewRecorder()

		web.ProjectsPageHandler(rec, req, *s, "all", "list")

		body := rec.Body.String()
		if !strings.Contains(body, "<title>") {
			t.Error("back-button to /projects must return full page")
		}
	})

	t.Run("project detail - HTMX partial has cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/project?id=0", nil)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.GetProjectHandler(rec, req, *s, "0", a)

		cacheControl := rec.Header().Get("Cache-Control")
		if cacheControl == "" {
			t.Error("HTMX partial for project detail must have Cache-Control header")
		}

		if !strings.Contains(cacheControl, "no-store") {
			t.Errorf("expected Cache-Control to contain 'no-store', got %q", cacheControl)
		}

		vary := rec.Header().Get("Vary")
		if !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain 'HX-Request', got %q", vary)
		}
	})

	t.Run("project detail - back button returns full page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/project?id=0", nil)
		rec := httptest.NewRecorder()

		web.GetProjectHandler(rec, req, *s, "0", a)

		body := rec.Body.String()
		if !strings.Contains(body, "<title>") {
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

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/reading-list", nil)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.ReadingListPageHandler(rec, req, *s, "all", "list")

		cacheControl := rec.Header().Get("Cache-Control")
		if cacheControl == "" {
			t.Error("HTMX partial for reading list must have Cache-Control header")
		}

		if !strings.Contains(cacheControl, "no-store") {
			t.Errorf("expected Cache-Control to contain 'no-store', got %q", cacheControl)
		}

		vary := rec.Header().Get("Vary")
		if !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain 'HX-Request', got %q", vary)
		}
	})

	t.Run("reading list - back button returns full page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/reading-list", nil)
		rec := httptest.NewRecorder()

		web.ReadingListPageHandler(rec, req, *s, "all", "list")

		body := rec.Body.String()
		if !strings.Contains(body, "<title>") {
			t.Error("back-button to /reading-list must return full page")
		}
	})

	t.Run("book detail - HTMX partial has cache headers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/book?id=0", nil)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.GetReadingListBook(rec, req, *s, "0", a)

		cacheControl := rec.Header().Get("Cache-Control")
		if cacheControl == "" {
			t.Error("HTMX partial for book detail must have Cache-Control header")
		}

		if !strings.Contains(cacheControl, "no-store") {
			t.Errorf("expected Cache-Control to contain 'no-store', got %q", cacheControl)
		}

		vary := rec.Header().Get("Vary")
		if !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain 'HX-Request', got %q", vary)
		}
	})

	t.Run("book detail - back button returns full page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/book?id=0", nil)
		rec := httptest.NewRecorder()

		web.GetReadingListBook(rec, req, *s, "0", a)

		body := rec.Body.String()
		if !strings.Contains(body, "<title>") {
			t.Error("back-button to /book must return full page")
		}
	})
}

// TestLettersBackButtonBehavior tests letters routes (requires authentication).
func TestLettersBackButtonBehavior(t *testing.T) {
	s := testSetup(t, context.Background())
	a, addAuthCookie := testAuthentication(t)

	t.Run("letters list - HTMX partial has cache headers", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/letters", nil)
		addAuthCookie(req)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.LettersPageHandler(rec, req, *s, "all", "list", a)

		cacheControl := rec.Header().Get("Cache-Control")
		if cacheControl == "" {
			t.Error("HTMX partial for letters must have Cache-Control header")
		}

		if !strings.Contains(cacheControl, "no-store") {
			t.Errorf("expected Cache-Control to contain 'no-store', got %q", cacheControl)
		}

		vary := rec.Header().Get("Vary")
		if !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain 'HX-Request', got %q", vary)
		}
	})

	t.Run("letters list - back button returns full page", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/letters", nil)
		addAuthCookie(req)

		rec := httptest.NewRecorder()

		web.LettersPageHandler(rec, req, *s, "all", "list", a)

		body := rec.Body.String()
		if !strings.Contains(body, "<title>") {
			t.Error("back-button to /letters must return full page")
		}
	})

	t.Run("letter detail - HTMX partial has cache headers", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/letter?id=0", nil)
		addAuthCookie(req)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.GetLetterHandler(rec, req, *s, "0", a)

		cacheControl := rec.Header().Get("Cache-Control")
		if cacheControl == "" {
			t.Error("HTMX partial for letter detail must have Cache-Control header")
		}

		if !strings.Contains(cacheControl, "no-store") {
			t.Errorf("expected Cache-Control to contain 'no-store', got %q", cacheControl)
		}

		vary := rec.Header().Get("Vary")
		if !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain 'HX-Request', got %q", vary)
		}
	})

	t.Run("letter detail - back button returns full page", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/letter?id=0", nil)
		addAuthCookie(req)

		rec := httptest.NewRecorder()

		web.GetLetterHandler(rec, req, *s, "0", a)

		body := rec.Body.String()
		if !strings.Contains(body, "<title>") {
			t.Error("back-button to /letter must return full page")
		}
	})
}
