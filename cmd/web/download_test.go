package web_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"timterests/cmd/web"
	"timterests/internal/storage"
)

func TestDownloadDocumentHandler(t *testing.T) {
	a, addAuthCookie := testAuthentication(t)

	t.Run("serves paired .md file when given a .yaml key", func(t *testing.T) {
		dir := t.TempDir()

		mdContent := "# Test\n\nbody content"
		mdPath := filepath.Join(dir, "articles", "test.md")

		if err := os.MkdirAll(filepath.Dir(mdPath), 0750); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		if err := os.WriteFile(mdPath, []byte(mdContent), 0600); err != nil {
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
