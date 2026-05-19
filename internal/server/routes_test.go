package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timterests/internal/server"
	"timterests/internal/storage"

	_ "github.com/mattn/go-sqlite3"
)

func TestRoutes(t *testing.T) {
	t.Parallel()
	t.Run("HelloWorldHandler returns expected response", func(t *testing.T) {
		t.Parallel()

		s := &server.Server{}

		svr := httptest.NewServer(http.HandlerFunc(s.HelloWorldHandler))
		defer svr.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, svr.URL, nil)
		if err != nil {
			t.Fatalf("error creating request. Err: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("error making request to svr. Err: %v", err)
		}

		defer func() {
			err := resp.Body.Close()
			if err != nil {
				t.Errorf("error closing response body: %v", err)
			}
		}()
		// Assertions
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status OK; got %v", resp.Status)
		}

		expected := "{\"message\":\"Hello World\"}"

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("error reading response body. Err: %v", err)
		}

		if expected != string(body) {
			t.Errorf("expected response body to be %v; got %v", expected, string(body))
		}
	})
}

func TestSecurityHeaders(t *testing.T) {
	setupHealthTestDB(t)

	s := &server.Server{
		Storage: &storage.Storage{
			UseS3:   false,
			BaseDir: t.TempDir(),
		},
	}

	svr := httptest.NewServer(s.RegisterRoutes())
	defer svr.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, svr.URL+"/health", nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("error making request: %v", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for name, want := range headers {
		got := resp.Header.Get(name)
		if got != want {
			t.Errorf("header %s: got %q, want %q", name, got, want)
		}
	}

	pp := resp.Header.Get("Permissions-Policy")
	if !strings.Contains(pp, "camera=()") {
		t.Errorf("Permissions-Policy missing camera=(), got %q", pp)
	}
}

func TestCORSPreflight(t *testing.T) {
	setupHealthTestDB(t)

	s := &server.Server{
		Storage: &storage.Storage{
			UseS3:   false,
			BaseDir: t.TempDir(),
		},
	}

	svr := httptest.NewServer(s.RegisterRoutes())
	defer svr.Close()

	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodOptions, svr.URL+"/health", nil,
	)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS preflight, got %d", resp.StatusCode)
	}

	corsHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS, PATCH",
	}

	for name, want := range corsHeaders {
		got := resp.Header.Get(name)
		if got != want {
			t.Errorf("header %s: got %q, want %q", name, got, want)
		}
	}
}

func TestHelloWorldContentType(t *testing.T) {
	t.Parallel()

	s := &server.Server{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, "/hello", nil,
	)

	s.HelloWorldHandler(rec, req)

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected JSON content type, got %q", ct)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Hello World") {
		t.Errorf("expected Hello World in body, got %q", body)
	}
}

func TestRegisterRoutesEndpoints(t *testing.T) {
	s := &server.Server{
		Storage: &storage.Storage{
			UseS3:   false,
			BaseDir: t.TempDir(),
		},
	}

	svr := httptest.NewServer(s.RegisterRoutes())
	defer svr.Close()

	endpoints := []string{
		"/",
		"/home",
		"/web",
		"/web/home",
		"/articles",
		"/projects",
		"/reading-list",
		"/login",
		"/sitemap.xml",
		"/about",
		"/writer",
		"/writer?type-id=invalid",
		"/admin",
		"/admin/documents",
		"/admin/users",
		"/admin/users/create",
		"/write",
		"/write/suggest",
		"/download",
		"/download/new",
		"/article",
		"/project",
		"/book",
		"/letter",
		"/letters",
	}

	for _, path := range endpoints {
		t.Run(path, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, svr.URL+path, nil)
			if err != nil {
				t.Fatalf("error creating request: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("error making request to %s: %v", path, err)
			}

			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == 0 {
				t.Errorf("expected non-zero status code for %s", path)
			}
		})
	}
}

func TestRecoveryMiddlewarePanic(t *testing.T) {
	s := &server.Server{
		Storage: &storage.Storage{
			UseS3:   false,
			BaseDir: t.TempDir(),
		},
		// auth is nil — /letters calls a.IsAuthenticated → nil deref → panic
	}

	svr := httptest.NewServer(s.RegisterRoutes())
	defer svr.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, svr.URL+"/letters", nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("error making request: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 from panic recovery, got %d", resp.StatusCode)
	}
}

func TestNewServer(t *testing.T) {
	t.Setenv("PORT", "18080")
	t.Setenv("SESSION_NAME", "test-session")

	svr := server.NewServer()
	if svr == nil {
		t.Fatal("expected non-nil server")
	}

	if svr.Addr != ":18080" {
		t.Errorf("expected addr :18080, got %s", svr.Addr)
	}
}
