package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timterests/internal/server"
	"timterests/internal/storage"
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
