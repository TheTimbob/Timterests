package errors_test

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	apperrors "timterests/internal/errors"
)

// TestInitLogger_CreatesFile verifies that InitLogger creates the log file.
func TestInitLogger_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "logs", "errors.log")

	if err := apperrors.InitLogger(logPath); err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	t.Cleanup(apperrors.CloseLogger)

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("expected log file to be created")
	}
}

// TestLogError_WritesJSONLine verifies a JSON line is appended to the log file.
func TestLogError_WritesJSONLine(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "logs", "errors.log")

	if err := apperrors.InitLogger(logPath); err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	t.Cleanup(apperrors.CloseLogger)

	appErr := apperrors.StorageReadFailed(errors.New("s3 timeout")).
		WithHandlerContext("test_handler", "test_action")

	apperrors.LogError(appErr)

	// Give the logger a moment to flush (it's synchronous, but be safe).
	time.Sleep(10 * time.Millisecond)

	// Close and re-read the file.
	apperrors.CloseLogger()

	f, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatal("expected at least one log line")
	}

	line := scanner.Text()
	if line == "" {
		t.Fatal("log line is empty")
	}

	// Should be valid JSON.
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("log line is not valid JSON: %v\nline: %s", err, line)
	}

	// Check required fields.
	for _, field := range []string{"timestamp", "severity", "code", "message"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("missing field %q in log entry", field)
		}
	}

	if entry["code"] != "STORAGE_READ_FAILED" {
		t.Errorf("expected code STORAGE_READ_FAILED, got %v", entry["code"])
	}

	if entry["severity"] != "ERROR" {
		t.Errorf("expected severity ERROR, got %v", entry["severity"])
	}

	if entry["handler"] != "test_handler" {
		t.Errorf("expected handler test_handler, got %v", entry["handler"])
	}

	if entry["underlying_error"] == nil || entry["underlying_error"] == "" {
		t.Error("expected underlying_error in log entry")
	}
}

// TestLogError_NilError_DoesNotPanic ensures LogError handles nil gracefully.
func TestLogError_NilError_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogError(nil) panicked: %v", r)
		}
	}()

	apperrors.LogError(nil)
}

// TestLogError_NoFileLogger_FallsBackToConsole ensures logging works without file setup.
func TestLogError_NoFileLogger_FallsBackToConsole(t *testing.T) {
	// Make sure there's no file logger initialized.
	apperrors.CloseLogger()

	appErr := apperrors.NotFound().WithHandlerContext("handler", "action")

	// Should not panic — only console output.
	apperrors.LogError(appErr)
}

// TestStackTrace_PresentForErrorSeverity checks that ERROR entries include stack_trace in log.
func TestStackTrace_PresentForErrorSeverity(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "logs", "errors.log")

	if err := apperrors.InitLogger(logPath); err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	t.Cleanup(apperrors.CloseLogger)

	appErr := apperrors.InternalServerError(errors.New("boom"))
	apperrors.LogError(appErr)

	apperrors.CloseLogger()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if st, ok := entry["stack_trace"]; !ok || st == "" {
		t.Error("expected stack_trace for ERROR severity")
	}
}

// TestStackTrace_AbsentForWarning checks that WARNING entries omit stack_trace.
func TestStackTrace_AbsentForWarning(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "logs", "errors.log")

	if err := apperrors.InitLogger(logPath); err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	t.Cleanup(apperrors.CloseLogger)

	appErr := apperrors.FileNotFound(errors.New("missing key"))
	apperrors.LogError(appErr)

	apperrors.CloseLogger()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if st, ok := entry["stack_trace"]; ok && st != "" {
		t.Error("expected no stack_trace for WARNING severity")
	}
}
