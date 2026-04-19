package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"timterests/cmd/web"
	"timterests/internal/auth"

	"github.com/PuerkitoBio/goquery"
)

func TestAdminPageHandler(t *testing.T) {
	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		a := auth.NewAuth("test-session-key-minimum-32-bytes")

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin", nil)
		rec := httptest.NewRecorder()

		web.AdminPageHandler(rec, req, a)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
		}

		if loc := rec.Header().Get("Location"); loc != "/login" {
			t.Errorf("expected redirect to /login, got %q", loc)
		}
	})

	t.Run("renders admin page when authenticated", func(t *testing.T) {
		a, addAuthCookie := testAuthentication(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin", nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)

		web.AdminPageHandler(rec, req, a)

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
