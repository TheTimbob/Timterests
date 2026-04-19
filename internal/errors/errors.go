package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityError    Severity = "ERROR"
	SeverityWarning  Severity = "WARNING"
	SeverityInfo     Severity = "INFO"
)

type AppError struct {
	Code       string
	Message    string
	Severity   Severity
	HTTPStatus int
	Handler    string
	Action     string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}

	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithHandler(handler, action string) *AppError {
	clone := *e

	clone.Handler = handler
	clone.Action = action

	return &clone
}

func (e *AppError) WithErr(err error) *AppError {
	clone := *e

	clone.Err = err

	return &clone
}

type errorDef struct {
	Code       string
	Message    string
	Severity   Severity
	HTTPStatus int
}

func getRegistry() map[string]errorDef {
	return map[string]errorDef{
		"INTERNAL_SERVER_ERROR": {
			"INTERNAL_SERVER_ERROR",
			"Something went wrong. Please try again later.",
			SeverityError, http.StatusInternalServerError,
		},
		"NOT_FOUND": {
			"NOT_FOUND",
			"The requested resource could not be found.",
			SeverityWarning, http.StatusNotFound,
		},
		"BAD_REQUEST": {
			"BAD_REQUEST",
			"The request contained invalid data.",
			SeverityWarning, http.StatusBadRequest,
		},
		"UNAUTHORIZED": {
			"UNAUTHORIZED",
			"Authentication required.",
			SeverityWarning, http.StatusUnauthorized,
		},
		"FORBIDDEN": {
			"FORBIDDEN",
			"You don't have permission to perform this action.",
			SeverityWarning, http.StatusForbidden,
		},
		"METHOD_NOT_ALLOWED": {
			"METHOD_NOT_ALLOWED",
			"Method not allowed.",
			SeverityWarning, http.StatusMethodNotAllowed,
		},
		"STORAGE_FAILED": {
			"STORAGE_FAILED",
			"Failed to access storage.",
			SeverityError, http.StatusInternalServerError,
		},
		"RENDER_FAILED": {
			"RENDER_FAILED",
			"Failed to render page.",
			SeverityError, http.StatusInternalServerError,
		},
		"PARSE_FORM_FAILED": {
			"PARSE_FORM_FAILED",
			"Failed to parse form data.",
			SeverityWarning, http.StatusBadRequest,
		},
		"LOGIN_FAILED": {
			"LOGIN_FAILED",
			"Incorrect email or password.",
			SeverityWarning, http.StatusUnauthorized,
		},
		"PANIC_RECOVERED": {
			"PANIC_RECOVERED",
			"An unexpected error occurred.",
			SeverityCritical, http.StatusInternalServerError,
		},
	}
}

func newFromDef(code string) *AppError {
	registry := getRegistry()

	def, ok := registry[code]
	if !ok {
		def = registry["INTERNAL_SERVER_ERROR"]
		def.Code = code
	}

	return &AppError{
		Code:       def.Code,
		Message:    def.Message,
		Severity:   def.Severity,
		HTTPStatus: def.HTTPStatus,
	}
}

func New(code string, err error) *AppError {
	return newFromDef(code).WithErr(err)
}

func InternalServerError(err error) *AppError { return New("INTERNAL_SERVER_ERROR", err) }
func NotFound(err error) *AppError            { return New("NOT_FOUND", err) }
func BadRequest(err error) *AppError          { return New("BAD_REQUEST", err) }
func Unauthorized(err error) *AppError        { return New("UNAUTHORIZED", err) }
func Forbidden() *AppError                    { return newFromDef("FORBIDDEN") }
func MethodNotAllowed() *AppError             { return newFromDef("METHOD_NOT_ALLOWED") }
func StorageFailed(err error) *AppError       { return New("STORAGE_FAILED", err) }
func RenderFailed(err error) *AppError        { return New("RENDER_FAILED", err) }
func ParseFormFailed(err error) *AppError     { return New("PARSE_FORM_FAILED", err) }
func LoginFailed(err error) *AppError         { return New("LOGIN_FAILED", err) }
func PanicRecovered(err error) *AppError      { return New("PANIC_RECOVERED", err) }

func Is(err, target error) bool { return errors.Is(err, target) }

func As(err error, target any) bool { return errors.As(err, target) }
