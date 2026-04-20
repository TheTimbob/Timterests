package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"timterests/cmd/web"
	"timterests/internal/auth"

	"github.com/PuerkitoBio/goquery"
)

func TestWriterPageHandler(t *testing.T) {
	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		s := testSetup(t, context.Background())
		a := auth.NewAuth("test-session-key-minimum-32-bytes")

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/writer", nil)
		rec := httptest.NewRecorder()

		web.WriterPageHandler(rec, req, *s, "articles", "", 0, a)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
		}

		if loc := rec.Header().Get("Location"); loc != "/login" {
			t.Errorf("expected redirect to /login, got %q", loc)
		}
	})

	t.Run("renders full writer page when authenticated", func(t *testing.T) {
		s := testSetup(t, context.Background())
		a, addAuthCookie := testAuthentication(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/writer", nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)

		web.WriterPageHandler(rec, req, *s, "articles", "", 0, a)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if doc.Find("title").Length() == 0 {
			t.Error("expected title element")
		}

		if doc.Find("#writer-container").Length() == 0 {
			t.Error("expected writer-container element")
		}
	})

	t.Run("renders partial form on HTMX request", func(t *testing.T) {
		s := testSetup(t, context.Background())
		a, addAuthCookie := testAuthentication(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/writer", nil)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		addAuthCookie(req)

		web.WriterPageHandler(rec, req, *s, "projects", "", 0, a)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if doc.Find("title").Length() > 0 {
			t.Error("expected no title element for partial render")
		}

		if doc.Find("#writer-form").Length() == 0 {
			t.Error("expected writer-form element in partial")
		}

		if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-store") {
			t.Errorf("expected Cache-Control to contain no-store, got %q", cc)
		}

		if vary := rec.Header().Get("Vary"); !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain HX-Request, got %q", vary)
		}
	})

	t.Run("renders different doc type fields", func(t *testing.T) {
		s := testSetup(t, context.Background())
		a, addAuthCookie := testAuthentication(t)

		docTypes := []struct {
			name     string
			fieldID  string
		}{
			{"articles", "date"},
			{"projects", "repository"},
			{"reading-list", "author"},
			{"letters", "occasion"},
		}

		for _, dt := range docTypes {
			t.Run(dt.name, func(t *testing.T) {
				req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/writer", nil)
				rec := httptest.NewRecorder()

				addAuthCookie(req)

				web.WriterPageHandler(rec, req, *s, dt.name, "", 0, a)

				doc, err := goquery.NewDocumentFromReader(rec.Body)
				if err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}

				if doc.Find("#" + dt.fieldID).Length() == 0 {
					t.Errorf("expected #%s field for doc type %q", dt.fieldID, dt.name)
				}
			})
		}
	})
}

func TestWriteDocumentHandler(t *testing.T) {
	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		s := testSetup(t, context.Background())
		a := auth.NewAuth("test-session-key-minimum-32-bytes")

		form := url.Values{}
		form.Set("document-type", "articles")
		form.Set("title", "Test")
		form.Set("subtitle", "Sub")
		form.Set("body", "Content")

		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodPost, "/write",
			strings.NewReader(form.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rec := httptest.NewRecorder()

		web.WriteDocumentHandler(rec, req, *s, a)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
		}
	})

	t.Run("rejects non-POST methods", func(t *testing.T) {
		s := testSetup(t, context.Background())
		a, addAuthCookie := testAuthentication(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/write", nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)

		web.WriteDocumentHandler(rec, req, *s, a)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", rec.Code)
		}
	})

	t.Run("writes document to local storage", func(t *testing.T) {
		a, addAuthCookie := testAuthentication(t)

		tmpDir := t.TempDir()

		s := testSetup(t, context.Background())
		s.BaseDir = tmpDir

		form := url.Values{}
		form.Set("document-type", "projects")
		form.Set("title", "My Test Project")
		form.Set("subtitle", "A subtitle")
		form.Set("preview", "Preview text")
		form.Set("body", "Project body content")
		form.Set("tags", "go,testing")
		form.Set("imagePath", "")
		form.Set("repository", "https://github.com/test/repo")

		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodPost, "/write",
			strings.NewReader(form.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		addAuthCookie(req)

		rec := httptest.NewRecorder()

		web.WriteDocumentHandler(rec, req, *s, a)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected redirect 303, got %d", rec.Code)
		}

		if loc := rec.Header().Get("Location"); loc != "/writer" {
			t.Errorf("expected redirect to /writer, got %q", loc)
		}
	})
}
