package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"timterests/cmd/web"
	"timterests/internal/storage"
)

// seedDocument writes a yaml/md pair into a throwaway storage dir so a delete
// test never touches the real testdata fixtures.
func seedDocument(t *testing.T, slug string) *storage.Storage {
	t.Helper()
	t.Setenv("USE_S3", "false")

	base := t.TempDir()
	dir := filepath.Join(base, "articles")

	err := os.MkdirAll(dir, 0750)
	if err != nil {
		t.Fatalf("failed to create %s: %v", dir, err)
	}

	for _, ext := range []string{".yaml", ".md"} {
		path := filepath.Join(dir, slug+ext)

		err = os.WriteFile(path, []byte("title: "+slug), 0600)
		if err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	return &storage.Storage{UseS3: false, BaseDir: base}
}

func deleteRequest(t *testing.T, key string) *http.Request {
	t.Helper()

	form := url.Values{}
	form.Set("key", key)

	req := httptest.NewRequestWithContext(
		context.Background(), http.MethodPost, "/admin/documents/delete",
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req
}

func requireExists(t *testing.T, path, reason string) {
	t.Helper()

	_, err := os.Stat(path)
	if err != nil {
		t.Error(reason)
	}
}

func TestDeleteDocumentHandler(t *testing.T) {
	a, addAuthCookie := testAuthentication(t)

	t.Run("removes both the yaml and md halves", func(t *testing.T) {
		s := seedDocument(t, "doomed")

		req := deleteRequest(t, "articles/doomed.yaml")
		addAuthCookie(req)

		web.DeleteDocumentHandler(httptest.NewRecorder(), req, *s, a)

		for _, ext := range []string{".yaml", ".md"} {
			path := filepath.Join(s.BaseDir, "articles", "doomed"+ext)

			_, err := os.Stat(path)
			if !os.IsNotExist(err) {
				t.Errorf("expected %s to be deleted", path)
			}
		}
	})

	// Deleting must not be reachable by a link, prefetch or stray GET.
	t.Run("rejects GET", func(t *testing.T) {
		s := seedDocument(t, "safe")

		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodGet,
			"/admin/documents/delete?key=articles/safe.yaml", nil,
		)
		addAuthCookie(req)

		web.DeleteDocumentHandler(httptest.NewRecorder(), req, *s, a)

		requireExists(t,
			filepath.Join(s.BaseDir, "articles", "safe.yaml"),
			"a GET must not delete anything",
		)
	})

	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		s := seedDocument(t, "protected")

		rec := httptest.NewRecorder()
		web.DeleteDocumentHandler(rec, deleteRequest(t, "articles/protected.yaml"), *s, a)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected redirect, got %d", rec.Code)
		}

		requireExists(t,
			filepath.Join(s.BaseDir, "articles", "protected.yaml"),
			"an unauthenticated request must not delete anything",
		)
	})

	// Keys outside the known content directories must be refused before they
	// reach the storage layer.
	t.Run("rejects keys outside content directories", func(t *testing.T) {
		for _, key := range []string{
			"../../etc/passwd",
			"secrets/thing.yaml",
			"articles/nested/deep.yaml",
			"articles/notyaml.txt",
			"",
		} {
			s := seedDocument(t, "bystander")

			rec := httptest.NewRecorder()
			req := deleteRequest(t, key)
			addAuthCookie(req)

			web.DeleteDocumentHandler(rec, req, *s, a)

			if rec.Code == http.StatusOK {
				t.Errorf("expected key %q to be rejected", key)
			}

			requireExists(t,
				filepath.Join(s.BaseDir, "articles", "bystander.yaml"),
				"key "+key+" deleted an unrelated document",
			)
		}
	})
}

// A half-written document — yaml present, md missing — must still be removable.
func TestDeleteDocumentToleratesMissingHalf(t *testing.T) {
	t.Setenv("USE_S3", "false")

	base := t.TempDir()
	dir := filepath.Join(base, "articles")

	err := os.MkdirAll(dir, 0750)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	yamlPath := filepath.Join(dir, "orphan.yaml")

	err = os.WriteFile(yamlPath, []byte("title: orphan"), 0600)
	if err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	s := &storage.Storage{UseS3: false, BaseDir: base}

	err = s.DeleteDocument(context.Background(), "articles/orphan.yaml")
	if err != nil {
		t.Fatalf("expected a missing body file to be tolerated, got %v", err)
	}

	_, err = os.Stat(yamlPath)
	if !os.IsNotExist(err) {
		t.Error("expected the yaml half to be deleted")
	}
}
