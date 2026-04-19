package errors

import (
	"errors"
	"log"
	"runtime/debug"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// LogError logs an AppError with consistent formatting and severity-based coloring.
func LogError(appErr *AppError) {
	if appErr == nil {
		return
	}

	color := severityColor(appErr.Severity)

	handler := appErr.Handler
	if handler == "" {
		handler = "-"
	}

	action := appErr.Action
	if action == "" {
		action = "-"
	}

	msg := appErr.Message
	if appErr.Err != nil {
		msg += ": " + appErr.Err.Error()
	}

	log.Printf("%s[%s] %s | handler=%s action=%s | %s%s",
		color, appErr.Severity, appErr.Code, handler, action, msg, colorReset)

	if appErr.Severity == SeverityCritical {
		log.Printf("%s[STACK] %s%s", color, string(debug.Stack()), colorReset)
	}
}

// Classify converts any error into an *AppError. If it's already an *AppError, it's
// returned as-is. Otherwise it's wrapped as INTERNAL_SERVER_ERROR.
func Classify(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return InternalServerError(err)
}

func severityColor(s Severity) string {
	switch s {
	case SeverityCritical, SeverityError:
		return colorRed
	case SeverityWarning:
		return colorYellow
	case SeverityInfo:
		return colorCyan
	default:
		return ""
	}
}
