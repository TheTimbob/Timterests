package errors

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ANSI color codes for console output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// logEntry is the JSON structure written to the log file.
type logEntry struct {
	Timestamp  string            `json:"timestamp"`
	Severity   string            `json:"severity"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Handler    string            `json:"handler,omitempty"`
	Action     string            `json:"action,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
	Underlying string            `json:"underlying_error,omitempty"`
	StackTrace string            `json:"stack_trace,omitempty"`
}

var (
	fileLogger *log.Logger
	logMu      sync.Mutex
	logFile    *os.File
)

// InitLogger opens (or creates) the log file and sets up the file logger.
// Call once at startup. If the logs/ directory doesn't exist, it is created.
func InitLogger(logPath string) error {
	logMu.Lock()
	defer logMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logFile = f
	fileLogger = log.New(f, "", 0) // timestamps handled in JSON

	return nil
}

// CloseLogger flushes and closes the log file.
func CloseLogger() {
	logMu.Lock()
	defer logMu.Unlock()

	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
		fileLogger = nil
	}
}

// LogError writes the AppError to both the log file (JSON) and the console (colored).
// It is safe to call before InitLogger; in that case only console output is produced.
func LogError(appErr *AppError) {
	if appErr == nil {
		return
	}

	// Copy the context map to avoid data races if the caller mutates it
	// concurrently while the entry is being JSON-marshaled.
	var ctxCopy map[string]string
	if len(appErr.Context) > 0 {
		ctxCopy = make(map[string]string, len(appErr.Context))
		for k, v := range appErr.Context {
			ctxCopy[k] = v
		}
	}

	entry := logEntry{
		Timestamp: appErr.Timestamp.Format(time.RFC3339),
		Severity:  string(appErr.Severity),
		Code:      appErr.Code,
		Message:   appErr.Message,
		Context:   ctxCopy,
	}

	if h, ok := appErr.Context["handler"]; ok {
		entry.Handler = h
	}

	if a, ok := appErr.Context["action"]; ok {
		entry.Action = a
	}

	if appErr.Err != nil {
		entry.Underlying = appErr.Err.Error()
	}

	// Capture stack trace for ERROR and CRITICAL.
	if appErr.Severity == SeverityError || appErr.Severity == SeverityCritical {
		entry.StackTrace = captureStack(3)
	}

	// Write JSON to file.
	logMu.Lock()
	if fileLogger != nil {
		data, err := json.Marshal(entry)
		if err == nil {
			fileLogger.Println(string(data))
		}
	}
	logMu.Unlock()

	// Write colored line to console.
	consoleLog(appErr, entry)
}

// consoleLog prints a human-readable colored line to stderr.
func consoleLog(appErr *AppError, entry logEntry) {
	color := severityColor(appErr.Severity)
	handler := entry.Handler
	if handler == "" {
		handler = "-"
	}

	action := entry.Action
	if action == "" {
		action = "-"
	}

	line := fmt.Sprintf("%s[%s] %s | %s | handler=%s action=%s",
		color,
		entry.Severity,
		entry.Code,
		entry.Message,
		handler,
		action,
	)

	if appErr.Err != nil {
		line += fmt.Sprintf(" | err=%v", appErr.Err)
	}

	line += colorReset

	log.Println(line)
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
		return colorWhite
	}
}

// captureStack returns a trimmed stack trace string, skipping skip frames.
func captureStack(skip int) string {
	const depth = 32
	var pcs [depth]uintptr

	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var sb strings.Builder

	for {
		frame, more := frames.Next()
		// Skip runtime internals.
		if strings.Contains(frame.Function, "runtime.") {
			if !more {
				break
			}

			continue
		}

		fmt.Fprintf(&sb, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)

		if !more {
			break
		}
	}

	return sb.String()
}
