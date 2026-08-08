package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timterests/cmd/web"
	"timterests/internal/server"
)

func TestStaticCacheMiddlewareSetsLifetime(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		cacheControl string
	}{
		{"image gets the long lifetime", "/assets/images/background.jpg", "public, max-age=604800"},
		{"stylesheet gets the short lifetime", "/assets/css/styles.css", "public, max-age=60"},
		{"script gets the short lifetime", "/assets/js/buttons.js", "public, max-age=60"},
	}

	handler := server.StaticCacheMiddleware(http.FileServer(http.FS(web.Files)))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil)

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 for %s, got %d", tt.path, rec.Code)
			}

			if got := rec.Header().Get("Cache-Control"); got != tt.cacheControl {
				t.Errorf("Cache-Control = %q, want %q", got, tt.cacheControl)
			}
		})
	}
}

func TestMaxBytesMiddlewareRejectsOversizedBody(t *testing.T) {
	s := &server.Server{}

	drainBody := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)

			return
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := s.MaxBytesMiddleware(drainBody)

	oversized := strings.NewReader(strings.Repeat("x", 11*1024*1024))

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", oversized)

	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Error("expected rejection for oversized body, but got 200")
	}
}
