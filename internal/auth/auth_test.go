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

func TestAuthPassword(t *testing.T) {
	t.Parallel()

	password := "Password1234!"

	t.Run("generate Hash", func(t *testing.T) {
		t.Parallel()

		hash, err := auth.GenerateHash(password)
		if err != nil {
			t.Fatalf("Failed to generate hash: %v", err)
		}

		if hash == "" {
			t.Errorf("Expected non-empty hash")
		}
	})

	t.Run("validate password is correct", func(t *testing.T) {
		t.Parallel()

		hash, err := auth.GenerateHash(password)
		if err != nil {
			t.Fatalf("Failed to generate hash: %v", err)
		}

		if !auth.ValidatePassword(password, hash) {
			t.Errorf("Expected password to be valid")
		}

		if auth.ValidatePassword("wrongpassword", hash) {
			t.Errorf("Expected password to be invalid")
		}
	})

	t.Run("generate Hash with empty password", func(t *testing.T) {
		t.Parallel()

		_, err := auth.GenerateHash("")
		if err == nil {
			t.Errorf("Expected error for empty password")
		}
	})
}

func TestNewAuth(t *testing.T) {
	t.Parallel()

	a := auth.NewAuth("test-session")
	if a == nil {
		t.Fatal("expected non-nil Auth")
	}
}

func TestIsAuthenticated(t *testing.T) {
	t.Parallel()

	t.Run("unauthenticated request returns false", func(t *testing.T) {
		t.Parallel()

		a := auth.NewAuth("test-session")
		r := newRequest()

		if a.IsAuthenticated(r) {
			t.Error("expected unauthenticated for fresh request")
		}
	})

	t.Run("authenticated after setting session", func(t *testing.T) {
		t.Parallel()

		a := auth.NewAuth("test-session")
		r := newRequest()
		w := httptest.NewRecorder()

		err := a.SetSessionValue(w, r, map[any]any{"email": "user@example.com"})
		if err != nil {
			t.Fatalf("SetSessionValue failed: %v", err)
		}

		if !a.IsAuthenticated(r) {
			t.Error("expected authenticated after setting email session value")
		}
	})

	t.Run("not authenticated with non-email session key", func(t *testing.T) {
		t.Parallel()

		a := auth.NewAuth("test-session")
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

	t.Run("sets and persists multiple values", func(t *testing.T) {
		t.Parallel()

		a := auth.NewAuth("test-session")
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

		if !a.IsAuthenticated(r) {
			t.Error("expected authenticated after setting email")
		}
	})
}
