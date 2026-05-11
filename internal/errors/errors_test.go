package errors_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	apperrors "timterests/internal/errors"
)

func captureLogOutput(fn func()) string {
	var buf bytes.Buffer

	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	fn()

	return buf.String()
}

func TestAppError(t *testing.T) {
	t.Parallel()

	t.Run("Error without underlying error", func(t *testing.T) {
		t.Parallel()

		err := apperrors.NotFound(nil)
		want := "[NOT_FOUND] The requested resource could not be found."

		if err.Error() != want {
			t.Fatalf("got %q, want %q", err.Error(), want)
		}
	})

	t.Run("Error with underlying error", func(t *testing.T) {
		t.Parallel()

		err := apperrors.StorageFailed(errors.New("disk full"))

		if err.Err == nil {
			t.Fatal("expected underlying error")
		}

		if err.HTTPStatus != http.StatusInternalServerError {
			t.Fatalf("got status %d, want %d", err.HTTPStatus, http.StatusInternalServerError)
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		t.Parallel()

		inner := errors.New("original")
		err := apperrors.InternalServerError(inner)

		if !errors.Is(err.Unwrap(), inner) {
			t.Fatal("Unwrap did not return the original error")
		}
	})

	t.Run("WithHandler creates independent copy", func(t *testing.T) {
		t.Parallel()

		err := apperrors.BadRequest(nil)
		withCtx := err.WithHandler("TestHandler", "doThing")

		if withCtx.Handler != "TestHandler" {
			t.Fatalf("got handler %q, want TestHandler", withCtx.Handler)
		}

		if err.Handler != "" {
			t.Fatal("original was mutated")
		}
	})

	t.Run("Unknown code falls back to INTERNAL_SERVER_ERROR", func(t *testing.T) {
		t.Parallel()

		err := apperrors.New("TOTALLY_MADE_UP", nil)

		if err.Code != "TOTALLY_MADE_UP" {
			t.Fatalf("got code %q, want TOTALLY_MADE_UP", err.Code)
		}

		if err.HTTPStatus != http.StatusInternalServerError {
			t.Fatalf("got status %d, want %d", err.HTTPStatus, http.StatusInternalServerError)
		}
	})
}

func TestClassify(t *testing.T) {
	t.Parallel()

	t.Run("classifies AppError as-is", func(t *testing.T) {
		t.Parallel()

		original := apperrors.NotFound(nil)
		classified := apperrors.Classify(original)

		if classified.Code != "NOT_FOUND" {
			t.Fatalf("got code %q, want NOT_FOUND", classified.Code)
		}
	})

	t.Run("wraps plain error as INTERNAL_SERVER_ERROR", func(t *testing.T) {
		t.Parallel()

		classified := apperrors.Classify(errors.New("something broke"))

		if classified.Code != "INTERNAL_SERVER_ERROR" {
			t.Fatalf("got code %q, want INTERNAL_SERVER_ERROR", classified.Code)
		}
	})
}

func TestSeverityConstructors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		err      *apperrors.AppError
		wantCode string
		wantSev  apperrors.Severity
	}{
		{"Unauthorized", apperrors.Unauthorized(nil), "UNAUTHORIZED", apperrors.SeverityWarning},
		{"Forbidden", apperrors.Forbidden(), "FORBIDDEN", apperrors.SeverityWarning},
		{"MethodNotAllowed", apperrors.MethodNotAllowed(), "METHOD_NOT_ALLOWED", apperrors.SeverityWarning},
		{"LoginFailed", apperrors.LoginFailed(nil), "LOGIN_FAILED", apperrors.SeverityWarning},
		{"PanicRecovered", apperrors.PanicRecovered(nil), "PANIC_RECOVERED", apperrors.SeverityCritical},
		{"RenderFailed", apperrors.RenderFailed(nil), "RENDER_FAILED", apperrors.SeverityError},
		{"ParseFormFailed", apperrors.ParseFormFailed(nil), "PARSE_FORM_FAILED", apperrors.SeverityWarning},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.err.Code != tc.wantCode {
				t.Fatalf("got code %q, want %q", tc.err.Code, tc.wantCode)
			}

			if tc.err.Severity != tc.wantSev {
				t.Fatalf("got severity %q, want %q", tc.err.Severity, tc.wantSev)
			}
		})
	}
}

func TestLogError(t *testing.T) {
	t.Run("logs error with handler context", func(t *testing.T) {
		err := apperrors.NotFound(errors.New("missing item"))
		err = err.WithHandler("ArticleHandler", "getArticle")

		output := captureLogOutput(func() { apperrors.LogError(err) })

		for _, want := range []string{
			"[WARNING]",
			"NOT_FOUND",
			"handler=ArticleHandler",
			"action=getArticle",
			"missing item",
		} {
			if !strings.Contains(output, want) {
				t.Errorf("expected log output to contain %q, got: %s", want, output)
			}
		}
	})

	t.Run("logs error without handler context uses dash fallback", func(t *testing.T) {
		err := apperrors.InternalServerError(errors.New("unexpected"))

		output := captureLogOutput(func() { apperrors.LogError(err) })

		for _, want := range []string{
			"[ERROR]",
			"INTERNAL_SERVER_ERROR",
			"handler=-",
			"action=-",
		} {
			if !strings.Contains(output, want) {
				t.Errorf("expected log output to contain %q, got: %s", want, output)
			}
		}
	})

	t.Run("logs critical error with stack trace", func(t *testing.T) {
		err := apperrors.PanicRecovered(errors.New("panic"))

		output := captureLogOutput(func() { apperrors.LogError(err) })

		for _, want := range []string{
			"[CRITICAL]",
			"PANIC_RECOVERED",
			"[STACK]",
			"goroutine",
		} {
			if !strings.Contains(output, want) {
				t.Errorf("expected log output to contain %q, got: %s", want, output)
			}
		}
	})

	t.Run("handles nil error gracefully", func(t *testing.T) {
		output := captureLogOutput(func() { apperrors.LogError(nil) })

		if output != "" {
			t.Errorf("expected no output for nil error, got: %s", output)
		}
	})

	t.Run("logs error without underlying error omits colon suffix", func(t *testing.T) {
		err := apperrors.Forbidden()

		output := captureLogOutput(func() { apperrors.LogError(err) })

		if !strings.Contains(output, "FORBIDDEN") {
			t.Errorf("expected FORBIDDEN in output, got: %s", output)
		}

		if strings.Contains(output, "You don't have permission to perform this action.:") {
			t.Error("message should not have trailing colon when there is no underlying error")
		}
	})

	t.Run("logs info severity with cyan color", func(t *testing.T) {
		err := &apperrors.AppError{
			Code:     "CUSTOM_INFO",
			Message:  "informational event",
			Severity: apperrors.SeverityInfo,
		}

		output := captureLogOutput(func() { apperrors.LogError(err) })

		for _, want := range []string{
			"[INFO]",
			"CUSTOM_INFO",
			"informational event",
			"\033[36m",
		} {
			if !strings.Contains(output, want) {
				t.Errorf("expected log output to contain %q, got: %s", want, output)
			}
		}
	})
}

func TestIsAndAs(t *testing.T) {
	t.Parallel()

	t.Run("Is delegates to errors.Is", func(t *testing.T) {
		t.Parallel()

		sentinel := errors.New("sentinel")
		wrapped := fmt.Errorf("wrapped: %w", sentinel)

		if !apperrors.Is(wrapped, sentinel) {
			t.Error("expected Is to find sentinel in wrapped error")
		}

		if apperrors.Is(wrapped, errors.New("other")) {
			t.Error("expected Is to return false for unrelated error")
		}
	})

	t.Run("As delegates to errors.As", func(t *testing.T) {
		t.Parallel()

		appErr := apperrors.NotFound(nil)

		var target *apperrors.AppError

		if !apperrors.As(appErr, &target) {
			t.Fatal("expected As to match AppError")
		}

		if target.Code != "NOT_FOUND" {
			t.Errorf("got code %q, want NOT_FOUND", target.Code)
		}
	})
}

func TestWithErr(t *testing.T) {
	t.Parallel()

	t.Run("creates independent copy with new error", func(t *testing.T) {
		t.Parallel()

		original := apperrors.NotFound(nil)
		inner := errors.New("specific reason")
		withErr := original.WithErr(inner)

		if !errors.Is(withErr.Err, inner) {
			t.Error("WithErr did not set the inner error")
		}

		if original.Err != nil {
			t.Error("original was mutated")
		}
	})
}
