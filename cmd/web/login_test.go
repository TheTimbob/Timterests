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

func TestLoginHandler(t *testing.T) {
	t.Run("renders login page on GET", func(t *testing.T) {
		a := auth.NewAuth("test-session-key-minimum-32-bytes")

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/login", nil)
		rec := httptest.NewRecorder()

		web.LoginHandler(rec, req, a)

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

		if doc.Find("form").Length() == 0 {
			t.Error("expected login form to be rendered")
		}
	})

	t.Run("renders partial login container on HTMX GET", func(t *testing.T) {
		a := auth.NewAuth("test-session-key-minimum-32-bytes")

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/login", nil)
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.LoginHandler(rec, req, a)

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

		if doc.Find("form").Length() == 0 {
			t.Error("expected login form to be rendered in partial")
		}
	})

	t.Run("returns 401 with error message on invalid credentials", func(t *testing.T) {
		a := auth.NewAuth("test-session-key-minimum-32-bytes")

		form := url.Values{}
		form.Set("email", "wrong@example.com")
		form.Set("password", "badpassword")

		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodPost, "/login",
			strings.NewReader(form.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rec := httptest.NewRecorder()

		web.LoginHandler(rec, req, a)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("HTMX POST with invalid credentials returns partial", func(t *testing.T) {
		a := auth.NewAuth("test-session-key-minimum-32-bytes")

		form := url.Values{}
		form.Set("email", "wrong@example.com")
		form.Set("password", "badpassword")

		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodPost, "/login",
			strings.NewReader(form.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Hx-Request", "true")

		rec := httptest.NewRecorder()

		web.LoginHandler(rec, req, a)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if doc.Find("title").Length() > 0 {
			t.Error("expected no title element for HTMX partial")
		}
	})
}
