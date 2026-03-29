package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apperrors "timterests/internal/errors"
)

// ---------------------------------------------------------------------------
// HandleError — response serialization
// ---------------------------------------------------------------------------

func TestHandleError_ReturnsJSONErrorPayload(t *testing.T) {
	w := httptest.NewRecorder()
	appErr := apperrors.StorageReadFailed(errors.New("db timeout"))

	HandleError(w, appErr, "test_handler", "test_action")

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}

	var body map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	errPayload, ok := body["error"]
	if !ok {
		t.Fatal("expected 'error' key in response")
	}

	if errPayload["code"] != "STORAGE_READ_FAILED" {
		t.Errorf("expected code STORAGE_READ_FAILED, got %v", errPayload["code"])
	}

	if errPayload["severity"] != "ERROR" {
		t.Errorf("expected severity ERROR, got %v", errPayload["severity"])
	}

	if errPayload["message"] == "" {
		t.Error("expected non-empty message")
	}

	// Verify NO stack trace or internal error in response.
	if _, hasStack := errPayload["stack_trace"]; hasStack {
		t.Error("response must not contain stack_trace")
	}

	if _, hasUnderlying := errPayload["underlying_error"]; hasUnderlying {
		t.Error("response must not contain underlying_error")
	}
}

func TestHandleError_NilError_ReturnsNil(t *testing.T) {
	w := httptest.NewRecorder()
	result := HandleError(w, nil, "handler", "action")

	if result != nil {
		t.Error("expected nil result for nil error")
	}
}

func TestHandleError_PlainError_WrapsAsInternalServerError(t *testing.T) {
	w := httptest.NewRecorder()
	HandleError(w, errors.New("something broke"), "handler", "action")

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}

	var body map[string]map[string]interface{}

	_ = json.NewDecoder(resp.Body).Decode(&body)

	if body["error"]["code"] != "INTERNAL_SERVER_ERROR" {
		t.Errorf("expected INTERNAL_SERVER_ERROR, got %v", body["error"]["code"])
	}
}

// ---------------------------------------------------------------------------
// HTTP status codes per error type
// ---------------------------------------------------------------------------

func TestHandleError_HTTPStatusCodes(t *testing.T) {
	cases := []struct {
		name       string
		err        *apperrors.AppError
		wantStatus int
	}{
		{"NotFound", apperrors.NotFound(), http.StatusNotFound},
		{"Unauthorized", apperrors.Unauthorized(), http.StatusForbidden},
		{"LoginFailed", apperrors.LoginFailed(errors.New("bad password")), http.StatusUnauthorized},
		{"InvalidInput", apperrors.InvalidInput(errors.New("missing field")), http.StatusBadRequest},
		{"SessionExpired", apperrors.SessionExpired(), http.StatusUnauthorized},
		{"StorageReadFailed", apperrors.StorageReadFailed(errors.New("timeout")), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleError(w, tc.err, "test", "test")

			if w.Code != tc.wantStatus {
				t.Errorf("want %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RecoveryMiddleware — panic catching
// ---------------------------------------------------------------------------

func TestRecoveryMiddleware_CatchesPanic(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went very wrong")
	})

	wrapped := RecoveryMiddleware(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should NOT panic.
	wrapped.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}

	var body map[string]map[string]interface{}

	_ = json.NewDecoder(resp.Body).Decode(&body)

	if body["error"]["code"] != "PANIC_RECOVERED" {
		t.Errorf("expected PANIC_RECOVERED, got %v", body["error"]["code"])
	}

	if body["error"]["severity"] != "CRITICAL" {
		t.Errorf("expected CRITICAL severity, got %v", body["error"]["severity"])
	}
}

func TestRecoveryMiddleware_NoPanic_PassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	wrapped := RecoveryMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Timestamp is present in response
// ---------------------------------------------------------------------------

func TestHandleError_ResponseContainsTimestamp(t *testing.T) {
	w := httptest.NewRecorder()
	HandleError(w, apperrors.Timeout(errors.New("took too long")), "h", "a")

	var body struct {
		Error struct {
			Timestamp time.Time `json:"timestamp"`
		} `json:"error"`
	}

	_ = json.NewDecoder(w.Body).Decode(&body)

	if body.Error.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp in response")
	}
}
