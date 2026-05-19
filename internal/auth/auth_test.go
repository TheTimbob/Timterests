package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"timterests/internal/auth"
	"timterests/internal/storage"

	_ "github.com/mattn/go-sqlite3"
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

	t.Run("authenticates user when multiple session values are set", func(t *testing.T) {
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

		r2 := newRequest()
		for _, c := range w.Result().Cookies() {
			r2.AddCookie(c)
		}

		if !a.IsAuthenticated(r2) {
			t.Error("expected authenticated after setting email")
		}
	})
}

func setupAuthDB(t *testing.T) {
	t.Helper()

	dbDir := filepath.Join(t.TempDir(), "database")

	err := os.MkdirAll(dbDir, 0750)
	if err != nil {
		t.Fatalf("failed to create database dir: %v", err)
	}

	dbPath := filepath.Join(dbDir, "timterests.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	err = storage.CreateUserTable(context.Background(), db)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	t.Chdir(filepath.Dir(dbDir))
}

func TestCreateUser(t *testing.T) {
	t.Run("creates user successfully", func(t *testing.T) {
		setupAuthDB(t)

		a := auth.NewAuth("test-session")
		ctx := context.Background()

		err := a.CreateUser(ctx, "John", "Doe", "john@example.com", "Pass123!")
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}
	})

	t.Run("rejects duplicate email", func(t *testing.T) {
		setupAuthDB(t)

		a := auth.NewAuth("test-session")
		ctx := context.Background()

		err := a.CreateUser(ctx, "A", "B", "dup@test.com", "Pass123!")
		if err != nil {
			t.Fatalf("first CreateUser failed: %v", err)
		}

		err = a.CreateUser(ctx, "C", "D", "dup@test.com", "Pass456!")
		if err == nil {
			t.Error("expected error for duplicate email")
		}
	})

	t.Run("rejects empty password", func(t *testing.T) {
		setupAuthDB(t)

		a := auth.NewAuth("test-session")

		err := a.CreateUser(context.Background(), "X", "Y", "x@test.com", "")
		if err == nil {
			t.Error("expected error for empty password")
		}
	})
}

func TestAuthenticate(t *testing.T) {
	t.Run("succeeds with correct credentials", func(t *testing.T) {
		setupAuthDB(t)

		a := auth.NewAuth("test-session")
		ctx := context.Background()

		err := a.CreateUser(ctx, "Auth", "User", "auth@test.com", "Correct123!")
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		w := httptest.NewRecorder()
		r := newRequest()

		ok, err := a.Authenticate(ctx, w, r, "auth@test.com", "Correct123!")
		if err != nil {
			t.Fatalf("Authenticate failed: %v", err)
		}

		if !ok {
			t.Error("expected authentication to succeed")
		}

		r2 := newRequest()
		for _, c := range w.Result().Cookies() {
			r2.AddCookie(c)
		}

		if !a.IsAuthenticated(r2) {
			t.Error("expected session to be set after authentication")
		}
	})

	t.Run("fails with wrong password", func(t *testing.T) {
		setupAuthDB(t)

		a := auth.NewAuth("test-session")
		ctx := context.Background()

		err := a.CreateUser(ctx, "Auth", "User", "wrong@test.com", "Correct123!")
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		w := httptest.NewRecorder()
		r := newRequest()

		ok, err := a.Authenticate(ctx, w, r, "wrong@test.com", "WrongPass!")
		if ok {
			t.Error("expected authentication to fail")
		}

		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("fails with nonexistent email", func(t *testing.T) {
		setupAuthDB(t)

		a := auth.NewAuth("test-session")

		w := httptest.NewRecorder()
		r := newRequest()

		ok, err := a.Authenticate(
			context.Background(), w, r, "nobody@test.com", "Pass123!",
		)
		if ok {
			t.Error("expected authentication to fail")
		}

		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}
