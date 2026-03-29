package errors_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	apperrors "timterests/internal/errors"
)

// ---------------------------------------------------------------------------
// Error construction
// ---------------------------------------------------------------------------

func TestNew_KnownCode(t *testing.T) {
	err := apperrors.New("STORAGE_READ_FAILED")

	if err.Code != "STORAGE_READ_FAILED" {
		t.Errorf("expected code STORAGE_READ_FAILED, got %q", err.Code)
	}

	if err.Severity != apperrors.SeverityError {
		t.Errorf("expected severity ERROR, got %q", err.Severity)
	}

	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected HTTP 500, got %d", err.HTTPStatus)
	}

	if err.Message == "" {
		t.Error("expected non-empty message")
	}

	if err.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestNew_UnknownCode_FallsBackToInternalServerError(t *testing.T) {
	err := apperrors.New("TOTALLY_MADE_UP_CODE")

	// Should still use the custom code (for debugging) but 500 severity/status.
	if err.Code != "TOTALLY_MADE_UP_CODE" {
		t.Errorf("expected original code preserved, got %q", err.Code)
	}

	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected HTTP 500 fallback, got %d", err.HTTPStatus)
	}
}

func TestWrap_AttachesUnderlyingError(t *testing.T) {
	underlying := errors.New("disk full")
	appErr := apperrors.Wrap("STORAGE_WRITE_FAILED", underlying)

	if appErr.Err == nil {
		t.Fatal("expected underlying error to be set")
	}

	if !errors.Is(appErr, underlying) {
		t.Error("expected errors.Is to find underlying error via Unwrap")
	}
}

// ---------------------------------------------------------------------------
// Severity / HTTP status classification
// ---------------------------------------------------------------------------

func TestSeverityMapping(t *testing.T) {
	cases := []struct {
		code     string
		severity apperrors.Severity
		status   int
	}{
		{"STORAGE_READ_FAILED", apperrors.SeverityError, http.StatusInternalServerError},
		{"STORAGE_WRITE_FAILED", apperrors.SeverityError, http.StatusInternalServerError},
		{"FILE_NOT_FOUND", apperrors.SeverityWarning, http.StatusNotFound},
		{"LOGIN_FAILED", apperrors.SeverityWarning, http.StatusUnauthorized},
		{"UNAUTHORIZED", apperrors.SeverityWarning, http.StatusForbidden},
		{"SESSION_EXPIRED", apperrors.SeverityCritical, http.StatusUnauthorized},
		{"INVALID_INPUT", apperrors.SeverityWarning, http.StatusBadRequest},
		{"PARSE_FORM_FAILED", apperrors.SeverityWarning, http.StatusBadRequest},
		{"TIMEOUT", apperrors.SeverityError, http.StatusGatewayTimeout},
		{"INTERNAL_SERVER_ERROR", apperrors.SeverityError, http.StatusInternalServerError},
		{"NOT_FOUND", apperrors.SeverityWarning, http.StatusNotFound},
		{"PANIC_RECOVERED", apperrors.SeverityCritical, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			appErr := apperrors.New(tc.code)

			if appErr.Severity != tc.severity {
				t.Errorf("severity: want %q, got %q", tc.severity, appErr.Severity)
			}

			if appErr.HTTPStatus != tc.status {
				t.Errorf("HTTPStatus: want %d, got %d", tc.status, appErr.HTTPStatus)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Context enrichment
// ---------------------------------------------------------------------------

func TestWithContext_DoesNotMutateOriginal(t *testing.T) {
	original := apperrors.New("INVALID_INPUT")
	enriched := original.WithContext("handler", "test_handler")

	if _, ok := original.Context["handler"]; ok {
		t.Error("WithContext should not mutate the original")
	}

	if enriched.Context["handler"] != "test_handler" {
		t.Errorf("enriched context missing handler, got %v", enriched.Context)
	}
}

func TestWithHandlerContext_SetsHandlerAndAction(t *testing.T) {
	appErr := apperrors.New("NOT_FOUND").WithHandlerContext("my_handler", "fetch_article")

	if appErr.Context["handler"] != "my_handler" {
		t.Errorf("expected handler=my_handler, got %q", appErr.Context["handler"])
	}

	if appErr.Context["action"] != "fetch_article" {
		t.Errorf("expected action=fetch_article, got %q", appErr.Context["action"])
	}

	// Timestamp should be refreshed.
	if time.Since(appErr.Timestamp) > 5*time.Second {
		t.Errorf("expected recent timestamp, got %v", appErr.Timestamp)
	}
}

// ---------------------------------------------------------------------------
// Convenience constructors
// ---------------------------------------------------------------------------

func TestConvenienceConstructors(t *testing.T) {
	underlying := errors.New("oops")

	cases := []struct {
		name    string
		err     *apperrors.AppError
		wantErr bool
	}{
		{"StorageReadFailed", apperrors.StorageReadFailed(underlying), true},
		{"StorageWriteFailed", apperrors.StorageWriteFailed(underlying), true},
		{"FileNotFound", apperrors.FileNotFound(underlying), true},
		{"LoginFailed", apperrors.LoginFailed(underlying), true},
		{"Unauthorized", apperrors.Unauthorized(), false},
		{"SessionExpired", apperrors.SessionExpired(), false},
		{"InvalidInput", apperrors.InvalidInput(underlying), true},
		{"ParseFormFailed", apperrors.ParseFormFailed(underlying), true},
		{"Timeout", apperrors.Timeout(underlying), true},
		{"InternalServerError", apperrors.InternalServerError(underlying), true},
		{"NotFound", apperrors.NotFound(), false},
		{"PanicRecovered", apperrors.PanicRecovered(), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatal("constructor returned nil")
			}

			if tc.wantErr && tc.err.Err == nil {
				t.Error("expected underlying error to be set")
			}

			if !tc.wantErr && tc.err.Err != nil {
				t.Errorf("expected no underlying error, got %v", tc.err.Err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error interface
// ---------------------------------------------------------------------------

func TestErrorString(t *testing.T) {
	appErr := apperrors.Wrap("FILE_NOT_FOUND", errors.New("no such key"))
	s := appErr.Error()

	if s == "" {
		t.Error("expected non-empty error string")
	}

	// Should contain the code.
	if !strings.Contains(s, "FILE_NOT_FOUND") {
		t.Errorf("error string should contain code, got: %q", s)
	}
}
