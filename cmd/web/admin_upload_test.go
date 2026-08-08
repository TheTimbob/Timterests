package web_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"timterests/cmd/web"
	"timterests/internal/storage"
)

// uploadRequest builds a multipart POST carrying the given file parts.
func uploadRequest(t *testing.T, docType string, parts map[string]string) *http.Request {
	t.Helper()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	err := writer.WriteField("document-type", docType)
	if err != nil {
		t.Fatalf("failed to write field: %v", err)
	}

	for field, spec := range parts {
		name, content, _ := strings.Cut(spec, "|")

		part, partErr := writer.CreateFormFile(field, name)
		if partErr != nil {
			t.Fatalf("failed to create part: %v", partErr)
		}

		_, partErr = part.Write([]byte(content))
		if partErr != nil {
			t.Fatalf("failed to write part: %v", partErr)
		}
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequestWithContext(
		context.Background(), http.MethodPost, "/admin/upload", &body,
	)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func uploadStorage(t *testing.T) *storage.Storage {
	t.Helper()
	t.Setenv("USE_S3", "false")

	return &storage.Storage{UseS3: false, BaseDir: t.TempDir()}
}

func TestUploadDocumentHandler(t *testing.T) {
	a, addAuthCookie := testAuthentication(t)

	validYAML := "post.yaml|title: A Post\ndate: 2026-01-01\n"
	validMD := "post.md|# A Post\n\nBody.\n"

	t.Run("writes both files when valid", func(t *testing.T) {
		s := uploadStorage(t)

		req := uploadRequest(t, "articles", map[string]string{
			"yaml-file": validYAML,
			"md-file":   validMD,
		})
		addAuthCookie(req)

		rec := httptest.NewRecorder()
		web.UploadDocumentHandler(rec, req, *s, a)

		for _, name := range []string{"post.yaml", "post.md"} {
			path := filepath.Join(s.BaseDir, "articles", name)

			_, err := os.Stat(path)
			if err != nil {
				t.Errorf("expected %s to be written", path)
			}
		}
	})

	// Validation must run before anything is written, so a rejected document
	// never lands half-uploaded.
	t.Run("rejects metadata missing a required field", func(t *testing.T) {
		s := uploadStorage(t)

		req := uploadRequest(t, "articles", map[string]string{
			"yaml-file": "post.yaml|title: No Date\n",
			"md-file":   validMD,
		})
		addAuthCookie(req)

		rec := httptest.NewRecorder()
		web.UploadDocumentHandler(rec, req, *s, a)

		if !strings.Contains(rec.Body.String(), "date") {
			t.Error("expected the response to name the missing field")
		}

		_, err := os.Stat(filepath.Join(s.BaseDir, "articles", "post.yaml"))
		if !os.IsNotExist(err) {
			t.Error("nothing should be written when validation fails")
		}
	})

	t.Run("rejects mismatched filenames", func(t *testing.T) {
		s := uploadStorage(t)

		req := uploadRequest(t, "articles", map[string]string{
			"yaml-file": validYAML,
			"md-file":   "different.md|# Other\n",
		})
		addAuthCookie(req)

		rec := httptest.NewRecorder()
		web.UploadDocumentHandler(rec, req, *s, a)

		if !strings.Contains(rec.Body.String(), "share a name") {
			t.Error("expected a filename mismatch to be reported")
		}
	})

	t.Run("rejects the wrong extension", func(t *testing.T) {
		s := uploadStorage(t)

		req := uploadRequest(t, "articles", map[string]string{
			"yaml-file": "post.txt|title: A Post\ndate: 2026-01-01\n",
			"md-file":   validMD,
		})
		addAuthCookie(req)

		rec := httptest.NewRecorder()
		web.UploadDocumentHandler(rec, req, *s, a)

		if !strings.Contains(rec.Body.String(), ".yaml") {
			t.Error("expected the extension to be rejected")
		}
	})

	t.Run("rejects an unknown document type", func(t *testing.T) {
		s := uploadStorage(t)

		req := uploadRequest(t, "nonsense", map[string]string{
			"yaml-file": validYAML,
			"md-file":   validMD,
		})
		addAuthCookie(req)

		rec := httptest.NewRecorder()
		web.UploadDocumentHandler(rec, req, *s, a)

		if !strings.Contains(rec.Body.String(), "Choose a document type") {
			t.Error("expected an unknown type to be rejected")
		}
	})

	// A traversal attempt in the filename must not escape the content directory.
	t.Run("strips directories from the filename", func(t *testing.T) {
		s := uploadStorage(t)

		req := uploadRequest(t, "articles", map[string]string{
			"yaml-file": "../../evil.yaml|title: Evil\ndate: 2026-01-01\n",
			"md-file":   "../../evil.md|# Evil\n",
		})
		addAuthCookie(req)

		rec := httptest.NewRecorder()
		web.UploadDocumentHandler(rec, req, *s, a)

		_, err := os.Stat(filepath.Join(s.BaseDir, "articles", "evil.yaml"))
		if err != nil {
			t.Error("expected the file to land inside the content directory")
		}

		_, err = os.Stat(filepath.Join(filepath.Dir(s.BaseDir), "evil.yaml"))
		if !os.IsNotExist(err) {
			t.Error("the upload escaped the content directory")
		}
	})

	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		s := uploadStorage(t)

		rec := httptest.NewRecorder()
		req := uploadRequest(t, "articles", map[string]string{
			"yaml-file": validYAML,
			"md-file":   validMD,
		})

		web.UploadDocumentHandler(rec, req, *s, a)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected a redirect, got %d", rec.Code)
		}
	})

	// A body large enough to arrive in several reads must be stored whole. A
	// single Read call would return only the first chunk and silently truncate.
	t.Run("stores a large body without truncating", func(t *testing.T) {
		s := uploadStorage(t)

		body := "# Big\n\n" + strings.Repeat("word ", 40000)

		req := uploadRequest(t, "articles", map[string]string{
			"yaml-file": validYAML,
			"md-file":   "post.md|" + body,
		})
		addAuthCookie(req)

		rec := httptest.NewRecorder()
		web.UploadDocumentHandler(rec, req, *s, a)

		written, err := os.ReadFile(filepath.Join(s.BaseDir, "articles", "post.md"))
		if err != nil {
			t.Fatalf("expected the body to be written: %v", err)
		}

		if len(written) != len(body) {
			t.Errorf("body truncated: wrote %d bytes, expected %d", len(written), len(body))
		}
	})
}

func TestUploadPageHandler(t *testing.T) {
	a, addAuthCookie := testAuthentication(t)

	t.Run("renders the form when authenticated", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/upload", nil)
		addAuthCookie(req)

		rec := httptest.NewRecorder()
		web.UploadPageHandler(rec, req, a)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		if !strings.Contains(rec.Body.String(), "yaml-file") {
			t.Error("expected the upload form to be rendered")
		}
	})

	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/upload", nil)

		rec := httptest.NewRecorder()
		web.UploadPageHandler(rec, req, a)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected a redirect, got %d", rec.Code)
		}
	})
}
