package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"timterests/internal/auth"
	"timterests/internal/storage"
)

func testSetup(t *testing.T, ctx context.Context) *storage.Storage {
	t.Helper()
	t.Setenv("USE_S3", "false")

	s, err := storage.NewStorage(ctx)
	if err != nil {
		t.Fatalf("failed to initialize storage: %v", err)
	}

	s.BaseDir = filepath.Join(s.BaseDir, "testdata")

	return s
}

// testAuthentication sets up authentication for all tests in a test function.
// It creates an Auth instance and returns both the instance and a function that adds
// the auth cookie to any request. Call this ONCE at the beginning of your test function,
// then use the returned Auth instance and cookie function for all sub-tests.
func testAuthentication(t *testing.T) (*auth.Auth, func(*http.Request)) {
	t.Helper()

	// Set up session name for testing (must be at least 32 bytes for AES)
	sessionName := "test-session-key-min-32-bytes!"
	t.Setenv("SESSION_NAME", sessionName)

	// Create a new Auth instance
	a := auth.NewAuth(sessionName)

	// Create a test request and recorder to capture the session cookie
	setupReq := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	setupRec := httptest.NewRecorder()

	// Set session value to simulate authenticated user
	err := a.SetSessionValue(setupRec, setupReq, map[any]any{"email": "test@example.com"})
	if err != nil {
		t.Fatalf("failed to set session: %v", err)
	}

	// Extract the session cookie
	cookies := setupRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no session cookie was set")
	}

	// Return the Auth instance and a function that adds the auth cookie to any request
	addAuthCookie := func(req *http.Request) {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}

	return a, addAuthCookie
}
