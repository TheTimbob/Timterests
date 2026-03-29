// Package errors provides a unified error handling system for Timterests.
// It defines AppError types with severity classification, an error registry
// of common errors, and utilities for wrapping errors with rich context.
package errors

import (
	"fmt"
	"net/http"
	"time"
)

// Severity represents the severity level of an error.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityError    Severity = "ERROR"
	SeverityWarning  Severity = "WARNING"
	SeverityInfo     Severity = "INFO"
)

// AppError is the central error type for Timterests. It carries everything
// needed for both logging (full context) and HTTP responses (clean metadata).
type AppError struct {
	// Code is a machine-readable identifier (e.g., "STORAGE_READ_FAILED").
	Code string `json:"code"`
	// Message is a user-friendly description (no internal details).
	Message string `json:"message"`
	// Severity classifies the impact: CRITICAL, ERROR, WARNING, INFO.
	Severity Severity `json:"severity"`
	// Timestamp is when the error occurred.
	Timestamp time.Time `json:"timestamp"`
	// Context holds structured key/value debug data (handler name, action, etc.).
	// This is logged but NOT exposed in HTTP responses.
	Context map[string]string `json:"-"`
	// Err is the underlying Go error (for logging only, never sent to clients).
	Err error `json:"-"`
	// HTTPStatus is the HTTP status code to return.
	HTTPStatus int `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}

	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap supports errors.Is / errors.As chaining.
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithContext returns a copy of the AppError with the given key/value added to Context.
func (e *AppError) WithContext(key, value string) *AppError {
	clone := *e
	clone.Context = make(map[string]string, len(e.Context)+1)
	for k, v := range e.Context {
		clone.Context[k] = v
	}

	clone.Context[key] = value

	return &clone
}

// WithErr returns a copy of the AppError with the underlying error set.
func (e *AppError) WithErr(err error) *AppError {
	clone := *e
	clone.Err = err

	return &clone
}

// WithHandlerContext returns a copy enriched with handler + action context.
func (e *AppError) WithHandlerContext(handler, action string) *AppError {
	clone := e.WithContext("handler", handler)
	clone = clone.WithContext("action", action)
	clone.Timestamp = time.Now().UTC()

	return clone
}

// errorDefinition is the template used to build registered errors.
type errorDefinition struct {
	Code       string
	Message    string
	Severity   Severity
	HTTPStatus int
}

// registry holds all known error definitions by code.
var registry = map[string]errorDefinition{
	// Storage errors
	"STORAGE_READ_FAILED": {
		Code:       "STORAGE_READ_FAILED",
		Message:    "We couldn't retrieve the requested data. Please try again.",
		Severity:   SeverityError,
		HTTPStatus: http.StatusInternalServerError,
	},
	"STORAGE_WRITE_FAILED": {
		Code:       "STORAGE_WRITE_FAILED",
		Message:    "We couldn't save your changes. Please try again.",
		Severity:   SeverityError,
		HTTPStatus: http.StatusInternalServerError,
	},
	"FILE_NOT_FOUND": {
		Code:       "FILE_NOT_FOUND",
		Message:    "The requested file or document could not be found.",
		Severity:   SeverityWarning,
		HTTPStatus: http.StatusNotFound,
	},
	// Auth errors
	"LOGIN_FAILED": {
		Code:       "LOGIN_FAILED",
		Message:    "Login failed. Please check your credentials and try again.",
		Severity:   SeverityWarning,
		HTTPStatus: http.StatusUnauthorized,
	},
	"UNAUTHORIZED": {
		Code:       "UNAUTHORIZED",
		Message:    "You don't have permission to perform this action.",
		Severity:   SeverityWarning,
		HTTPStatus: http.StatusForbidden,
	},
	"SESSION_EXPIRED": {
		Code:       "SESSION_EXPIRED",
		Message:    "Your session has expired. Please log in again.",
		Severity:   SeverityCritical,
		HTTPStatus: http.StatusUnauthorized,
	},
	// Input errors
	"INVALID_INPUT": {
		Code:       "INVALID_INPUT",
		Message:    "The request contained invalid or missing data.",
		Severity:   SeverityWarning,
		HTTPStatus: http.StatusBadRequest,
	},
	"PARSE_FORM_FAILED": {
		Code:       "PARSE_FORM_FAILED",
		Message:    "We couldn't read the form data. Please try again.",
		Severity:   SeverityWarning,
		HTTPStatus: http.StatusBadRequest,
	},
	// Network / timeout errors
	"TIMEOUT": {
		Code:       "TIMEOUT",
		Message:    "The request took too long. Please try again.",
		Severity:   SeverityError,
		HTTPStatus: http.StatusGatewayTimeout,
	},
	// Generic / catch-all
	"INTERNAL_SERVER_ERROR": {
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "Something went wrong on our end. Please try again later.",
		Severity:   SeverityError,
		HTTPStatus: http.StatusInternalServerError,
	},
	"NOT_FOUND": {
		Code:       "NOT_FOUND",
		Message:    "The requested resource could not be found.",
		Severity:   SeverityWarning,
		HTTPStatus: http.StatusNotFound,
	},
	"PANIC_RECOVERED": {
		Code:       "PANIC_RECOVERED",
		Message:    "An unexpected error occurred. Our team has been notified.",
		Severity:   SeverityCritical,
		HTTPStatus: http.StatusInternalServerError,
	},
}

// New creates a fresh AppError from the registry by code.
// If the code is not found, it falls back to INTERNAL_SERVER_ERROR.
func New(code string) *AppError {
	def, ok := registry[code]
	if !ok {
		def = registry["INTERNAL_SERVER_ERROR"]
		def.Code = code // preserve the unknown code for debugging
	}

	return &AppError{
		Code:       def.Code,
		Message:    def.Message,
		Severity:   def.Severity,
		HTTPStatus: def.HTTPStatus,
		Timestamp:  time.Now().UTC(),
		Context:    make(map[string]string),
	}
}

// Wrap creates a registered AppError and attaches the underlying Go error.
func Wrap(code string, err error) *AppError {
	return New(code).WithErr(err)
}

// Pre-built error constructors for convenience.

func StorageReadFailed(err error) *AppError  { return Wrap("STORAGE_READ_FAILED", err) }
func StorageWriteFailed(err error) *AppError { return Wrap("STORAGE_WRITE_FAILED", err) }
func FileNotFound(err error) *AppError       { return Wrap("FILE_NOT_FOUND", err) }
func LoginFailed(err error) *AppError        { return Wrap("LOGIN_FAILED", err) }
func Unauthorized() *AppError                { return New("UNAUTHORIZED") }
func SessionExpired() *AppError              { return New("SESSION_EXPIRED") }
func InvalidInput(err error) *AppError       { return Wrap("INVALID_INPUT", err) }
func ParseFormFailed(err error) *AppError    { return Wrap("PARSE_FORM_FAILED", err) }
func Timeout(err error) *AppError            { return Wrap("TIMEOUT", err) }
func InternalServerError(err error) *AppError {
	return Wrap("INTERNAL_SERVER_ERROR", err)
}
func NotFound() *AppError       { return New("NOT_FOUND") }
func PanicRecovered() *AppError { return New("PANIC_RECOVERED") }
