package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"timterests/internal/server"
	"timterests/internal/storage"
)

func TestHealthHandlerOK(t *testing.T) {
	setupHealthTestDB(t)

	s := &server.Server{
		Storage: &storage.Storage{
			UseS3:   false,
			BaseDir: t.TempDir(),
		},
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result map[string]any

	err := json.Unmarshal(rec.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", result["status"])
	}

	if result["ts"] == nil || result["ts"] == "" {
		t.Error("expected 'ts' field to be present and non-empty")
	}

	checks, ok := result["checks"].(map[string]any)
	if !ok {
		t.Fatal("expected 'checks' field to be an object")
	}

	if checks["storage"] != "ok" {
		t.Errorf("expected storage 'ok', got %q", checks["storage"])
	}

	if checks["database"] != "ok" {
		t.Errorf("expected database 'ok', got %q", checks["database"])
	}
}

func TestHealthHandlerDegraded(t *testing.T) {
	t.Chdir(t.TempDir())

	s := &server.Server{
		Storage: &storage.Storage{
			UseS3:   false,
			BaseDir: t.TempDir(),
		},
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.HealthHandler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}

	var result map[string]any

	err := json.Unmarshal(rec.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if result["status"] != "degraded" {
		t.Errorf("expected status 'degraded', got %q", result["status"])
	}
}
