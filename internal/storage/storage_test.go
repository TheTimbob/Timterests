package storage_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"timterests/internal/model"
	"timterests/internal/storage"

	_ "github.com/mattn/go-sqlite3"
)

func TestStorage(t *testing.T) {
	t.Parallel()

	t.Run("create new storage instance", func(t *testing.T) {
		s, err := storage.NewStorage(t.Context())
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if s.UseS3 != false {
			t.Errorf("Expected UseS3 to be false, got %v", s.UseS3)
		}

		if s.BucketName != "" {
			t.Errorf("Expected empty BucketName for local storage, got %v", s.BucketName)
		}
		// For local storage, BaseDir should be set
		if s.BaseDir == "" {
			t.Errorf("Expected BaseDir to be set, got empty string")
		}
	})

	t.Run("decode yaml document", func(t *testing.T) {
		t.Parallel()

		var (
			document model.Document
			filename = "document.yaml"
		)

		fs := &fstest.MapFS{
			filename: getYAMLDocument(),
		}

		file, err := fs.Open(filename)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		err = storage.DecodeFile(file, &document)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedTitle := "Test Document"
		if document.Title != expectedTitle {
			t.Errorf("Expected title '%s', got %v", expectedTitle, document.Title)
		}

		expectedPreview := "A brief preview."
		if document.Preview != expectedPreview {
			t.Errorf("Expected preview %q, got %q", expectedPreview, document.Preview)
		}
	})

	t.Run("write markdown document creates yaml and md files", func(t *testing.T) {
		t.Parallel()

		formData := map[string]any{
			"title":    "Test Document",
			"subtitle": "Test Subtitle",
			"preview":  "A brief preview.",
			"body":     "Test Body content.",
			"tags":     []string{"test", "document"},
		}

		tempDir := t.TempDir()
		yamlPath := tempDir + "/test-document.yaml"
		mdPath := tempDir + "/test-document.md"

		err := storage.WriteMarkdownDocument(yamlPath, mdPath, formData)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		_, err = os.Stat(yamlPath)
		if os.IsNotExist(err) {
			t.Fatalf("YAML file was not created")
		}

		_, err = os.Stat(mdPath)
		if os.IsNotExist(err) {
			t.Fatalf("Markdown file was not created")
		}
	})

	t.Run("write and re-read yaml metadata", func(t *testing.T) {
		t.Parallel()

		formData := map[string]any{
			"title":    "Round Trip",
			"subtitle": "Subtitle",
			"preview":  "Preview text.",
			"body":     "Body content here.",
		}

		tempDir := t.TempDir()
		yamlPath := tempDir + "/round-trip.yaml"
		mdPath := tempDir + "/round-trip.md"

		err := storage.WriteMarkdownDocument(yamlPath, mdPath, formData)
		if err != nil {
			t.Fatalf("Expected no error writing, got %v", err)
		}

		file, err := os.Open(yamlPath)
		if err != nil {
			t.Fatalf("Expected no error opening yaml, got %v", err)
		}
		defer file.Close()

		var doc model.Document

		err = storage.DecodeFile(file, &doc)
		if err != nil {
			t.Fatalf("Expected no error decoding, got %v", err)
		}

		if doc.Title != "Round Trip" {
			t.Errorf("Expected title 'Round Trip', got %q", doc.Title)
		}

		if doc.Preview != "Preview text." {
			t.Errorf("Expected preview 'Preview text.', got %q", doc.Preview)
		}
	})

	t.Run("write and re-read markdown body", func(t *testing.T) {
		t.Parallel()

		formData := map[string]any{
			"title":    "Body Test",
			"subtitle": "Sub",
			"body":     "The actual body content.",
		}

		tempDir := t.TempDir()
		yamlPath := tempDir + "/body-test.yaml"
		mdPath := tempDir + "/body-test.md"

		err := storage.WriteMarkdownDocument(yamlPath, mdPath, formData)
		if err != nil {
			t.Fatalf("Expected no error writing, got %v", err)
		}

		content, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("Expected no error reading md file, got %v", err)
		}

		if !strings.Contains(string(content), "The actual body content.") {
			t.Errorf("Markdown file does not contain expected body, got: %s", string(content))
		}

		if !strings.HasPrefix(string(content), "# Body Test") {
			t.Errorf("Markdown file should start with title header, got: %s", string(content))
		}
	})
}

func TestGetPromptContent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	promptsDir := t.TempDir()

	docTypes := []string{"articles", "projects", "reading-list", "letters"}
	for _, docType := range docTypes {
		promptFile := filepath.Join(promptsDir, docType+".txt")

		content := "test prompt for " + docType

		err := os.WriteFile(promptFile, []byte(content), 0600)
		if err != nil {
			t.Fatalf("failed to write prompt file for %s: %v", docType, err)
		}
	}

	s := &storage.Storage{
		UseS3:      false,
		BucketName: "",
		BaseDir:    t.TempDir(),
		PromptsDir: promptsDir,
		S3Client:   nil,
	}

	tests := []struct {
		name          string
		docType       string
		expectError   bool
		expectContent string
	}{
		{
			name:          "articles",
			docType:       "articles",
			expectError:   false,
			expectContent: "test prompt for articles",
		},
		{
			name:          "projects",
			docType:       "projects",
			expectError:   false,
			expectContent: "test prompt for projects",
		},
		{
			name:          "reading-list",
			docType:       "reading-list",
			expectError:   false,
			expectContent: "test prompt for reading-list",
		},
		{
			name:          "letters",
			docType:       "letters",
			expectError:   false,
			expectContent: "test prompt for letters",
		},
		{
			name:        "invalid docType",
			docType:     "invalid",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content, err := s.GetPromptContent(ctx, tc.docType)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error for docType %q, got nil", tc.docType)
				}

				return
			}

			if err != nil {
				t.Errorf("GetPromptContent failed for %q: %v", tc.docType, err)

				return
			}

			if content != tc.expectContent {
				t.Errorf("expected content %q, got %q", tc.expectContent, content)
			}
		})
	}
}

func TestHealthOK(t *testing.T) {
	setupHealthDB(t)

	s := &storage.Storage{
		UseS3:   false,
		BaseDir: t.TempDir(),
	}

	result := s.Health()

	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", result.Status)
	}

	if result.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}

	_, parseErr := time.Parse(time.RFC3339, result.Timestamp)
	if parseErr != nil {
		t.Errorf("timestamp is not valid RFC3339: %q", result.Timestamp)
	}

	if result.Checks["storage"] != "ok" {
		t.Errorf("expected storage check 'ok', got %q", result.Checks["storage"])
	}

	if result.Checks["database"] != "ok" {
		t.Errorf("expected database check 'ok', got %q", result.Checks["database"])
	}

	if !result.Healthy() {
		t.Error("expected Healthy() to return true")
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal health result: %v", err)
	}

	var parsed map[string]any

	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("failed to unmarshal health JSON: %v", err)
	}

	for _, key := range []string{"status", "ts", "checks"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON response", key)
		}
	}
}

func TestHealthDegradedDBDown(t *testing.T) {
	// No database directory → GetDB will fail → database check returns error
	t.Chdir(t.TempDir())

	s := &storage.Storage{
		UseS3:   false,
		BaseDir: t.TempDir(),
	}

	result := s.Health()

	if result.Status != "degraded" {
		t.Errorf("expected status 'degraded', got %q", result.Status)
	}

	if result.Checks["storage"] != "ok" {
		t.Errorf("expected storage 'ok', got %q", result.Checks["storage"])
	}

	if result.Checks["database"] == "ok" {
		t.Error("expected database check to report error, got 'ok'")
	}

	if result.Healthy() {
		t.Error("expected Healthy() to return false for degraded status")
	}
}

func TestHealthDegradedStorageDown(t *testing.T) {
	setupHealthDB(t)

	s := &storage.Storage{
		UseS3:   false,
		BaseDir: "/nonexistent/path/that/does/not/exist",
	}

	result := s.Health()

	if result.Status != "degraded" {
		t.Errorf("expected status 'degraded', got %q", result.Status)
	}

	if result.Checks["storage"] == "ok" {
		t.Error("expected storage check to report error, got 'ok'")
	}

	if result.Checks["database"] != "ok" {
		t.Errorf("expected database 'ok', got %q", result.Checks["database"])
	}

	if result.Healthy() {
		t.Error("expected Healthy() to return false for degraded status")
	}
}

func TestFormatFileSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"small file", 512, "512 B"},
		{"exactly 1KB boundary", 1024, "1.0 KB"},
		{"kilobytes", 2048, "2.0 KB"},
		{"fractional KB", 1536, "1.5 KB"},
		{"large file", 1048576, "1024.0 KB"},
		{"just under 1KB", 1023, "1023 B"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := storage.FormatFileSize(tc.size)
			if result != tc.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tc.size, result, tc.expected)
			}
		})
	}
}

func setupHealthDB(t *testing.T) {
	t.Helper()

	dbDir := filepath.Join(t.TempDir(), "database")

	err := os.MkdirAll(dbDir, 0750)
	if err != nil {
		t.Fatalf("failed to create database dir: %v", err)
	}

	dbPath := filepath.Join(dbDir, "timterests.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	defer func() {
		_ = db.Close()
	}()

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	t.Chdir(filepath.Dir(dbDir))
}

func TestInitDB(t *testing.T) {
	t.Run("creates users table when missing", func(t *testing.T) {
		dbDir := filepath.Join(t.TempDir(), "database")

		err := os.MkdirAll(dbDir, 0750)
		if err != nil {
			t.Fatalf("failed to create database dir: %v", err)
		}

		t.Chdir(filepath.Dir(dbDir))

		err = storage.InitDB(context.Background())
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		db, err := sql.Open("sqlite3", filepath.Join(dbDir, "timterests.db"))
		if err != nil {
			t.Fatalf("failed to open db: %v", err)
		}
		defer db.Close()

		var count int

		err = db.QueryRowContext(
			context.Background(),
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'",
		).Scan(&count)
		if err != nil {
			t.Fatalf("failed to query sqlite_master: %v", err)
		}

		if count != 1 {
			t.Errorf("expected users table to exist, got count %d", count)
		}
	})

	t.Run("is idempotent when table already exists", func(t *testing.T) {
		dbDir := filepath.Join(t.TempDir(), "database")

		err := os.MkdirAll(dbDir, 0750)
		if err != nil {
			t.Fatalf("failed to create database dir: %v", err)
		}

		t.Chdir(filepath.Dir(dbDir))

		err = storage.InitDB(context.Background())
		if err != nil {
			t.Fatalf("first InitDB failed: %v", err)
		}

		err = storage.InitDB(context.Background())
		if err != nil {
			t.Fatalf("second InitDB failed: %v", err)
		}
	})

	t.Run("returns error when no database directory", func(t *testing.T) {
		t.Chdir(t.TempDir())

		err := storage.InitDB(context.Background())
		if err == nil {
			t.Error("expected error when database directory is missing")
		}
	})
}

func TestCreateUserTable(t *testing.T) {
	t.Run("creates table and allows insert", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")

		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			t.Fatalf("failed to open db: %v", err)
		}
		defer db.Close()

		err = storage.CreateUserTable(context.Background(), db)
		if err != nil {
			t.Fatalf("CreateUserTable failed: %v", err)
		}

		_, err = db.ExecContext(
			context.Background(),
			"INSERT INTO users (first_name, last_name, email, password) VALUES (?, ?, ?, ?)",
			"Test", "User", "test@example.com", "hash",
		)
		if err != nil {
			t.Fatalf("insert into users failed: %v", err)
		}
	})
}

func TestLocalPath(t *testing.T) {
	t.Parallel()

	t.Run("valid relative path", func(t *testing.T) {
		t.Parallel()

		result, err := storage.LocalPath("/base", "subdir/file.yaml")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result != "/base/subdir/file.yaml" {
			t.Errorf("expected /base/subdir/file.yaml, got %s", result)
		}
	})

	t.Run("path traversal is rejected", func(t *testing.T) {
		t.Parallel()

		_, err := storage.LocalPath("/base", "../etc/passwd")
		if err == nil {
			t.Error("expected error for path traversal, got nil")
		}
	})

	t.Run("absolute filename is rejected", func(t *testing.T) {
		t.Parallel()

		_, err := storage.LocalPath("/base", "/absolute/path")
		if err == nil {
			t.Error("expected error for absolute filename, got nil")
		}
	})
}

func TestListObjects(t *testing.T) {
	t.Parallel()

	t.Run("lists files from existing directory", func(t *testing.T) {
		t.Parallel()

		baseDir := t.TempDir()
		dir := filepath.Join(baseDir, "articles")

		err := os.MkdirAll(dir, 0750)
		if err != nil {
			t.Fatalf("failed to create articles dir: %v", err)
		}

		err = os.WriteFile(filepath.Join(dir, "one.yaml"), []byte("title: One"), 0600)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		err = os.WriteFile(filepath.Join(dir, "two.yaml"), []byte("title: Two"), 0600)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		s := &storage.Storage{UseS3: false, BaseDir: baseDir}

		objects, err := s.ListObjects(context.Background(), "articles")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(objects) != 2 {
			t.Errorf("expected 2 objects, got %d", len(objects))
		}
	})

	t.Run("creates directory and returns empty list", func(t *testing.T) {
		t.Parallel()

		s := &storage.Storage{UseS3: false, BaseDir: t.TempDir()}

		objects, err := s.ListObjects(context.Background(), "newdir")
		if err != nil {
			t.Fatalf("expected no error for new dir, got %v", err)
		}

		if len(objects) != 0 {
			t.Errorf("expected 0 objects, got %d", len(objects))
		}
	})

	t.Run("ignores subdirectories", func(t *testing.T) {
		t.Parallel()

		baseDir := t.TempDir()
		dir := filepath.Join(baseDir, "articles")

		err := os.MkdirAll(filepath.Join(dir, "subdir"), 0750)
		if err != nil {
			t.Fatalf("failed to create subdirs: %v", err)
		}

		err = os.WriteFile(filepath.Join(dir, "file.yaml"), []byte("title: File"), 0600)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		s := &storage.Storage{UseS3: false, BaseDir: baseDir}

		objects, err := s.ListObjects(context.Background(), "articles")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(objects) != 1 {
			t.Errorf("expected 1 object (subdir ignored), got %d", len(objects))
		}
	})
}

func TestGetFile(t *testing.T) {
	t.Parallel()

	t.Run("opens existing local file", func(t *testing.T) {
		t.Parallel()

		baseDir := t.TempDir()
		dir := filepath.Join(baseDir, "articles")

		err := os.MkdirAll(dir, 0750)
		if err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		err = os.WriteFile(filepath.Join(dir, "test.yaml"), []byte("title: Test"), 0600)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		s := &storage.Storage{UseS3: false, BaseDir: baseDir}

		file, err := s.GetFile(context.Background(), "articles/test.yaml")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		defer file.Close()
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		t.Parallel()

		s := &storage.Storage{UseS3: false, BaseDir: t.TempDir()}

		_, err := s.GetFile(context.Background(), "articles/missing.yaml")
		if err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})
}

func TestGetPreparedFile(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	dir := filepath.Join(baseDir, "articles")

	err := os.MkdirAll(dir, 0750)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(dir, "test.yaml"), []byte("title: Prepared Article\nsubtitle: Sub"), 0600)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	s := &storage.Storage{UseS3: false, BaseDir: baseDir}

	var doc model.Document

	err = s.GetPreparedFile(context.Background(), "articles/test.yaml", &doc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if doc.Title != "Prepared Article" {
		t.Errorf("expected title 'Prepared Article', got %q", doc.Title)
	}
}

func TestGetRawFile(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	dir := filepath.Join(baseDir, "articles")

	err := os.MkdirAll(dir, 0750)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(dir, "raw.yaml"), []byte("title: Raw Article"), 0600)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	s := &storage.Storage{UseS3: false, BaseDir: baseDir}

	var doc model.Document

	err = s.GetRawFile(context.Background(), "articles/raw.yaml", &doc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if doc.Title != "Raw Article" {
		t.Errorf("expected title 'Raw Article', got %q", doc.Title)
	}
}

func TestGetImage(t *testing.T) {
	t.Parallel()

	t.Run("local mode returns web path", func(t *testing.T) {
		t.Parallel()

		s := &storage.Storage{UseS3: false, BaseDir: t.TempDir()}

		result, err := s.GetImage(context.Background(), "images/photo.jpg")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expected := "/storage/images/photo.jpg"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("returns error for path traversal", func(t *testing.T) {
		t.Parallel()

		s := &storage.Storage{UseS3: false, BaseDir: t.TempDir()}

		_, err := s.GetImage(context.Background(), "../etc/passwd")
		if err == nil {
			t.Error("expected error for path traversal, got nil")
		}
	})
}

func TestGetDocumentBodyRaw(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	dir := filepath.Join(baseDir, "articles")

	err := os.MkdirAll(dir, 0750)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	mdContent := "# Test Article\n\nThis is the body."

	err = os.WriteFile(filepath.Join(dir, "test.md"), []byte(mdContent), 0600)
	if err != nil {
		t.Fatalf("failed to write md file: %v", err)
	}

	s := &storage.Storage{UseS3: false, BaseDir: baseDir}

	raw, err := s.GetDocumentBodyRaw(context.Background(), "articles/test.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if raw != mdContent {
		t.Errorf("expected %q, got %q", mdContent, raw)
	}
}

func TestGetDocumentBody(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	dir := filepath.Join(baseDir, "articles")

	err := os.MkdirAll(dir, 0750)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(dir, "test.md"), []byte("# Title\n\nBody text."), 0600)
	if err != nil {
		t.Fatalf("failed to write md file: %v", err)
	}

	s := &storage.Storage{UseS3: false, BaseDir: baseDir}

	html, err := s.GetDocumentBody(context.Background(), "articles/test.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(html, "<h1") {
		t.Errorf("expected HTML with h1 tag, got %q", html)
	}
}

func TestDownloadS3FileLocalMode(t *testing.T) {
	t.Parallel()

	s := &storage.Storage{UseS3: false, BaseDir: t.TempDir()}

	err := s.DownloadS3File(context.Background(), "articles/test.yaml")
	if err != nil {
		t.Errorf("expected no error in local mode, got %v", err)
	}
}

func getYAMLDocument() *fstest.MapFile {
	return &fstest.MapFile{
		Data: []byte(
			"title: Test Document\nsubtitle: Test Subtitle\n" +
				"preview: A brief preview.\ntags:\n  - test\n  - document\n",
		),
	}
}
