package web_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timterests/cmd/web"
	apperrors "timterests/internal/errors"
)

func TestHandleError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "not found renders 404 page",
			err:        apperrors.NotFound(nil),
			wantStatus: http.StatusNotFound,
			wantBody:   "404",
		},
		{
			name:       "bad request renders 400 page",
			err:        apperrors.BadRequest(errors.New("invalid")),
			wantStatus: http.StatusBadRequest,
			wantBody:   "400",
		},
		{
			name:       "internal error renders 500 page",
			err:        apperrors.InternalServerError(errors.New("db down")),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "500",
		},
		{
			name:       "unauthorized renders 401 page",
			err:        apperrors.Unauthorized(nil),
			wantStatus: http.StatusUnauthorized,
			wantBody:   "401",
		},
		{
			name:       "generic error classified as 500",
			err:        errors.New("something unexpected"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(
				context.Background(), http.MethodGet, "/test", nil,
			)
			rec := httptest.NewRecorder()

			web.HandleError(rec, req, tt.err, "TestHandler", "testAction")

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			body := rec.Body.String()
			if !strings.Contains(body, tt.wantBody) {
				t.Errorf("expected body to contain %q", tt.wantBody)
			}
		})
	}
}

func TestHandleErrorRendersHTML(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet, "/missing", nil,
	)
	rec := httptest.NewRecorder()

	web.HandleError(rec, req, apperrors.NotFound(nil), "handler", "action")

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %q", ct)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Back to Home") {
		t.Error("error page missing 'Back to Home' link")
	}
}
