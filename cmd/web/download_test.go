package web_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"timterests/cmd/web"
	"timterests/internal/auth"
	"timterests/internal/storage"
)

func TestDownloadNewDocumentHandler(t *testing.T) {
	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		a := auth.NewAuth("test-session-key-minimum-32-bytes")

		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/download/new", nil)
		rec := httptest.NewRecorder()

		web.DownloadNewDocumentHandler(rec, req, a)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("returns 400 when title is missing", func(t *testing.T) {
		a, addAuthCookie := testAuthentication(t)

		form := url.Values{}
		form.Set("subtitle", "Sub")
		form.Set("body", "Content")

		req := httptest.NewRequestWithContext(
			t.Context(), http.MethodPost, "/download/new",
			strings.NewReader(form.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		addAuthCookie(req)

		rec := httptest.NewRecorder()

		web.DownloadNewDocumentHandler(rec, req, a)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("serves markdown file with correct headers", func(t *testing.T) {
		a, addAuthCookie := testAuthentication(t)

		form := url.Values{}
		form.Set("title", "My Document")
		form.Set("subtitle", "A subtitle")
		form.Set("body", "Body content here")

		req := httptest.NewRequestWithContext(
			t.Context(), http.MethodPost, "/download/new",
			strings.NewReader(form.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		addAuthCookie(req)

		rec := httptest.NewRecorder()

		web.DownloadNewDocumentHandler(rec, req, a)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "text/markdown" {
			t.Errorf("expected Content-Type text/markdown, got %q", ct)
		}

		body := rec.Body.String()

		if !strings.Contains(body, "# My Document") {
			t.Error("expected markdown to contain title header")
		}

		if !strings.Contains(body, "## A subtitle") {
			t.Error("expected markdown to contain subtitle header")
		}

		if !strings.Contains(body, "Body content here") {
			t.Error("expected markdown to contain body content")
		}
	})
}

func TestDownloadDocumentHandler(t *testing.T) {
	a, addAuthCookie := testAuthentication(t)

	t.Run("serves paired .md file when given a .yaml key", func(t *testing.T) {
		dir := t.TempDir()

		mdContent := "# Test\n\nbody content"
		mdPath := filepath.Join(dir, "articles", "test.md")

		err := os.MkdirAll(filepath.Dir(mdPath), 0750)
		if err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		err = os.WriteFile(mdPath, []byte(mdContent), 0600)
		if err != nil {
			t.Fatalf("failed to write md file: %v", err)
		}

		s := storage.Storage{UseS3: false, BaseDir: dir}

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/download?key=articles/test.yaml", nil)
		addAuthCookie(req)

		rec := httptest.NewRecorder()

		web.DownloadDocumentHandler(rec, req, s, "articles/test.yaml", a)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "text/markdown" {
			t.Errorf("expected Content-Type text/markdown, got %q", ct)
		}

		if cd := rec.Header().Get("Content-Disposition"); cd == "" {
			t.Error("expected Content-Disposition header to be set")
		}
	})

	t.Run("returns 400 for missing key", func(t *testing.T) {
		s := storage.Storage{UseS3: false, BaseDir: t.TempDir()}

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/download", nil)
		addAuthCookie(req)

		rec := httptest.NewRecorder()

		web.DownloadDocumentHandler(rec, req, s, "", a)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("returns 401 for unauthenticated request", func(t *testing.T) {
		s := storage.Storage{UseS3: false, BaseDir: t.TempDir()}

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/download?key=articles/test.yaml", nil)
		rec := httptest.NewRecorder()

		web.DownloadDocumentHandler(rec, req, s, "articles/test.yaml", a)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})
}
