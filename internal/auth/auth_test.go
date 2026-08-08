package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"timterests/internal/auth"
)

func newRequest() *http.Request {
	return httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
}

func TestNewAuth(t *testing.T) {
	t.Parallel()

	a := auth.NewAuth("test-session", testSessionKey)
	if a == nil {
		t.Fatal("expected non-nil Auth")
	}
}

func TestIsAuthenticated(t *testing.T) {
	t.Parallel()

	t.Run("unauthenticated request returns false", func(t *testing.T) {
		t.Parallel()

		a := auth.NewAuth("test-session", testSessionKey)
		r := newRequest()

		if a.IsAuthenticated(r) {
			t.Error("expected unauthenticated for fresh request")
		}
	})

	t.Run("authenticated after setting session", func(t *testing.T) {
		t.Parallel()

		a := auth.NewAuth("test-session", testSessionKey)
		r := newRequest()
		w := httptest.NewRecorder()

		err := a.SetSessionValue(w, r, map[any]any{"email": "user@example.com"})
		if err != nil {
			t.Fatalf("SetSessionValue failed: %v", err)
		}

		// Use a fresh request with the saved cookie to avoid gorilla's request-level cache,
		// which would make the assertion pass even if session.Save were never called.
		r2 := newRequest()
		for _, c := range w.Result().Cookies() {
			r2.AddCookie(c)
		}

		if !a.IsAuthenticated(r2) {
			t.Error("expected authenticated after setting email session value")
		}
	})

	t.Run("not authenticated with non-email session key", func(t *testing.T) {
		t.Parallel()

		a := auth.NewAuth("test-session", testSessionKey)
		r := newRequest()
		w := httptest.NewRecorder()

		err := a.SetSessionValue(w, r, map[any]any{"role": "admin"})
		if err != nil {
			t.Fatalf("SetSessionValue failed: %v", err)
		}

		if a.IsAuthenticated(r) {
			t.Error("expected unauthenticated when only non-email key is set")
		}
	})
}

func TestSetSessionValue(t *testing.T) {
	t.Parallel()

	t.Run("authenticates user when multiple session values are set", func(t *testing.T) {
		t.Parallel()

		a := auth.NewAuth("test-session", testSessionKey)
		r := newRequest()
		w := httptest.NewRecorder()

		values := map[any]any{
			"email": "test@example.com",
			"role":  "admin",
		}

		err := a.SetSessionValue(w, r, values)
		if err != nil {
			t.Fatalf("SetSessionValue failed: %v", err)
		}

		r2 := newRequest()
		for _, c := range w.Result().Cookies() {
			r2.AddCookie(c)
		}

		if !a.IsAuthenticated(r2) {
			t.Error("expected authenticated after setting email")
		}
	})
}
