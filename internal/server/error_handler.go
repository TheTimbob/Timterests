// Package server — error_handler.go
// Provides HTTP-level error handling: the HandleError helper and panic-recovery middleware.
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	apperrors "timterests/internal/errors"
)

// errorResponse is the JSON shape returned to clients.
// It deliberately omits stack traces and internal details.
type errorResponse struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
}

// HandleError classifies err as an AppError (or wraps it as INTERNAL_SERVER_ERROR),
// logs it with full context, and writes the appropriate JSON response.
// Returns the resulting *AppError so callers can inspect it if needed.
func HandleError(w http.ResponseWriter, err error, handlerName, action string) *apperrors.AppError {
	if err == nil {
		return nil
	}

	var appErr *apperrors.AppError

	// If it's already an AppError, enrich it with handler context.
	if ae, ok := err.(*apperrors.AppError); ok {
		appErr = ae.WithHandlerContext(handlerName, action)
	} else {
		// Unknown error — wrap as INTERNAL_SERVER_ERROR.
		appErr = apperrors.InternalServerError(err).WithHandlerContext(handlerName, action)
	}

	// Log to file + console.
	apperrors.LogError(appErr)

	// Write JSON response (safe for clients: no internal details).
	writeErrorResponse(w, appErr)

	return appErr
}

// writeErrorResponse sets headers and writes the JSON error payload.
func writeErrorResponse(w http.ResponseWriter, appErr *apperrors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)

	resp := errorResponse{
		Error: errorPayload{
			Code:      appErr.Code,
			Message:   appErr.Message,
			Severity:  string(appErr.Severity),
			Timestamp: appErr.Timestamp,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		// Last-resort fallback — should never happen.
		log.Printf("error_handler: failed to marshal error response: %v", err)
		http.Error(w, `{"error":{"code":"MARSHAL_FAILED","message":"Internal error"}}`, http.StatusInternalServerError)

		return
	}

	_, err = w.Write(data)
	if err != nil {
		log.Printf("error_handler: failed to write error response: %v", err)
	}
}

// RecoveryMiddleware wraps an http.Handler and converts panics into
// PANIC_RECOVERED AppErrors so the server never crashes on handler panics.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				panicErr := fmt.Errorf("panic: %v", rec)
				appErr := apperrors.PanicRecovered().
					WithErr(panicErr).
					WithHandlerContext(r.URL.Path, r.Method)

				apperrors.LogError(appErr)
				writeErrorResponse(w, appErr)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
