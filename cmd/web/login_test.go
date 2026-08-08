package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timterests/cmd/web"

	"github.com/PuerkitoBio/goquery"
)

// enableOIDC sets the Cognito variables the login page checks before offering
// the sign-in link.
func enableOIDC(t *testing.T) {
	t.Helper()

	t.Setenv("COGNITO_DOMAIN", "example.auth.us-east-2.amazoncognito.com")
	t.Setenv("COGNITO_USER_POOL_ID", "us-east-2_abc123")
	t.Setenv("COGNITO_CLIENT_ID", "client-id")
	t.Setenv("COGNITO_CLIENT_SECRET", "client-secret")
	t.Setenv("SITE_URL", "http://localhost:8080")
}

func TestLoginHandler(t *testing.T) {
	t.Run("renders login page on GET", func(t *testing.T) {
		enableOIDC(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/login", nil)
		rec := httptest.NewRecorder()

		web.LoginHandler(rec, req)

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

		if doc.Find("a[href='/auth/login']").Length() == 0 {
			t.Error("expected the Google sign-in link to be rendered")
		}
	})

	t.Run("renders partial login container on HTMX GET", func(t *testing.T) {
		enableOIDC(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/login", nil)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.LoginHandler(rec, req)

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

		if doc.Find("a[href='/auth/login']").Length() == 0 {
			t.Error("expected the Google sign-in link in partial")
		}

		if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-store") {
			t.Errorf("expected Cache-Control to contain no-store, got %q", cc)
		}

		if vary := rec.Header().Get("Vary"); !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain HX-Request, got %q", vary)
		}
	})
}

func TestLoginHandlerWithoutOIDC(t *testing.T) {
	t.Setenv("COGNITO_DOMAIN", "")
	t.Setenv("COGNITO_USER_POOL_ID", "")
	t.Setenv("COGNITO_CLIENT_ID", "")
	t.Setenv("COGNITO_CLIENT_SECRET", "")
	t.Setenv("SITE_URL", "")

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	web.LoginHandler(rec, req)

	doc, err := goquery.NewDocumentFromReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if doc.Find("a[href='/auth/login']").Length() != 0 {
		t.Error("expected no sign-in link when Cognito is unconfigured")
	}
}
