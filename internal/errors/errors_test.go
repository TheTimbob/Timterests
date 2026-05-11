package errors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	apperrors "timterests/internal/errors"
)

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
	t.Parallel()

	t.Run("logs error with handler context", func(t *testing.T) {
		t.Parallel()

		err := apperrors.NotFound(errors.New("missing item"))
		err = err.WithHandler("ArticleHandler", "getArticle")

		apperrors.LogError(err)
	})

	t.Run("logs error without handler context", func(t *testing.T) {
		t.Parallel()

		err := apperrors.InternalServerError(errors.New("unexpected"))

		apperrors.LogError(err)
	})

	t.Run("logs critical error with stack trace", func(t *testing.T) {
		t.Parallel()

		err := apperrors.PanicRecovered(errors.New("panic"))

		apperrors.LogError(err)
	})

	t.Run("handles nil error gracefully", func(t *testing.T) {
		t.Parallel()

		apperrors.LogError(nil)
	})

	t.Run("logs error without underlying error", func(t *testing.T) {
		t.Parallel()

		err := apperrors.Forbidden()

		apperrors.LogError(err)
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
