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

func TestAdminUsersPageHandler(t *testing.T) {
	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		a := auth.NewAuth("test-session-key-minimum-32-bytes")

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/users", nil)
		rec := httptest.NewRecorder()

		web.AdminUsersPageHandler(rec, req, a)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
		}

		if loc := rec.Header().Get("Location"); loc != "/login" {
			t.Errorf("expected redirect to /login, got %q", loc)
		}
	})

	t.Run("renders full page when authenticated", func(t *testing.T) {
		a, addAuthCookie := testAuthentication(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/users", nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)

		web.AdminUsersPageHandler(rec, req, a)

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

		if doc.Find(`[id="admin-users-container"]`).Length() == 0 {
			t.Error("expected admin-users-container, but it wasn't found")
		}

		if doc.Find("h1.category-title").Text() != "Create User" {
			t.Errorf("expected page title 'Create User', got %q", doc.Find("h1.category-title").Text())
		}
	})

	t.Run("renders partial on HTMX request", func(t *testing.T) {
		a, addAuthCookie := testAuthentication(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/users", nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)
		req.Header.Set("Hx-Request", "true")

		web.AdminUsersPageHandler(rec, req, a)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if doc.Find("title").Length() > 0 {
			t.Error("expected no title element for partial render, but found one")
		}

		if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-store") {
			t.Errorf("expected Cache-Control to contain no-store, got %q", cc)
		}
	})
}

func TestCreateUserHandler(t *testing.T) {
	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		a := auth.NewAuth("test-session-key-minimum-32-bytes")

		form := url.Values{}
		form.Set("first_name", "Test")
		form.Set("last_name", "User")
		form.Set("email", "test@example.com")
		form.Set("password", "password123")

		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodPost, "/admin/users/create",
			strings.NewReader(form.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rec := httptest.NewRecorder()

		web.CreateUserHandler(rec, req, a)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
		}
	})

	t.Run("rejects non-POST method", func(t *testing.T) {
		a, addAuthCookie := testAuthentication(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/users/create", nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)

		web.CreateUserHandler(rec, req, a)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
	})

	t.Run("returns error when fields are missing", func(t *testing.T) {
		a, addAuthCookie := testAuthentication(t)

		form := url.Values{}
		form.Set("first_name", "Test")

		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodPost, "/admin/users/create",
			strings.NewReader(form.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rec := httptest.NewRecorder()

		addAuthCookie(req)

		web.CreateUserHandler(rec, req, a)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		errorMsg := doc.Find(".error-message").Text()
		if errorMsg == "" {
			t.Error("expected error message, but none found")
		}
	})
}
