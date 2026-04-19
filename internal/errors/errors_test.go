package errors_test

import (
	"errors"
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
		{"Forbidden", apperrors.Forbidden(), "FORBIDDEN", apperrors.SeverityWarning},
		{"MethodNotAllowed", apperrors.MethodNotAllowed(), "METHOD_NOT_ALLOWED", apperrors.SeverityWarning},
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
