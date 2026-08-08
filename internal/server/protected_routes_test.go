package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timterests/internal/server"
)

// protectedRoutes are the paths that must never serve content to a signed-out
// visitor. Adding an admin route without wrapping it in requireAuth should make
// this fail rather than silently expose it.
func protectedRoutes() []string {
	return []string{
		"/admin",
		"/admin/documents",
		"/admin/documents/delete",
		"/admin/upload",
		"/writer",
		"/write",
		"/download",
		"/download/new",
		"/letters",
		"/letter",
	}
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()

	// No working-directory isolation here: NewServer resolves the storage
	// directory from the project root, which an empty temp dir does not have.
	t.Setenv("PORT", "18082")
	t.Setenv("USE_S3", "false")
	t.Setenv("SESSION_NAME", "test-session")
	t.Setenv("SESSION_KEY", "test-signing-key-at-least-32-chars!!")

	return server.NewServer().Handler
}

func TestProtectedRoutesRedirectWhenSignedOut(t *testing.T) {
	handler := newTestServer(t)

	for _, route := range protectedRoutes() {
		t.Run(route, func(t *testing.T) {
			for _, method := range []string{http.MethodGet, http.MethodPost} {
				req := httptest.NewRequestWithContext(t.Context(), method, route, nil)
				rec := httptest.NewRecorder()

				handler.ServeHTTP(rec, req)

				if rec.Code != http.StatusSeeOther {
					t.Errorf("%s %s: expected a redirect, got %d", method, route, rec.Code)
				}

				if loc := rec.Header().Get("Location"); loc != "/login" {
					t.Errorf("%s %s: expected redirect to /login, got %q", method, route, loc)
				}
			}
		})
	}
}

// The public pages must stay reachable — the guard should not have crept onto
// anything a visitor needs.
func TestPublicRoutesStayReachable(t *testing.T) {
	handler := newTestServer(t)

	for _, route := range []string{"/articles", "/projects", "/reading-list", "/login", "/health"} {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, route, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code == http.StatusSeeOther {
				if loc := rec.Header().Get("Location"); strings.Contains(loc, "/login") {
					t.Errorf("%s should not require signing in", route)
				}
			}
		})
	}
}
