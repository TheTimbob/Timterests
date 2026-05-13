package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timterests/internal/server"
)

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
