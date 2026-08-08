package storage_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"timterests/internal/model"
	"timterests/internal/storage"
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

	t.Run("write markdown document escapes html", func(t *testing.T) {
		t.Parallel()

		formData := map[string]any{
			"title":    `<script>alert("title")</script>`,
			"subtitle": `<img src=x onerror=alert("subtitle")>`,
			"body":     `<script>alert("body")</script>`,
		}

		tempDir := t.TempDir()
		yamlPath := tempDir + "/escaped.yaml"
		mdPath := tempDir + "/escaped.md"

		err := storage.WriteMarkdownDocument(yamlPath, mdPath, formData)
		if err != nil {
			t.Fatalf("Expected no error writing, got %v", err)
		}

		content, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("Expected no error reading md file, got %v", err)
		}

		body := string(content)

		if strings.Contains(body, "<script>") || strings.Contains(body, "<img") {
			t.Fatalf("Expected markdown file to escape raw HTML, got: %s", body)
		}

		if !strings.Contains(body, "&lt;script&gt;alert(&#34;body&#34;)&lt;/script&gt;") {
			t.Errorf("Expected escaped body HTML, got: %s", body)
		}
	})
}

func TestHealthOK(t *testing.T) {
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

func TestHealthDegradedStorageDown(t *testing.T) {
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

func TestLocalPath(t *testing.T) {
	t.Parallel()

	t.Run("valid relative path", func(t *testing.T) {
		t.Parallel()

		result, err := storage.LocalPath("/base", "articles/test.md")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Build the expectation the same way LocalPath does, so the assertion
		// does not depend on the host's path separator.
		expected := filepath.Join("/base", "articles/test.md")
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("rejects absolute path", func(t *testing.T) {
		t.Parallel()

		_, err := storage.LocalPath("/base", "/etc/passwd")
		if err == nil {
			t.Error("expected error for absolute path traversal")
		}
	})

	t.Run("rejects parent directory traversal", func(t *testing.T) {
		t.Parallel()

		_, err := storage.LocalPath("/base", "../../../etc/passwd")
		if err == nil {
			t.Error("expected error for parent directory traversal")
		}
	})

	t.Run("rejects empty filename", func(t *testing.T) {
		t.Parallel()

		_, err := storage.LocalPath("/base", "")
		if err == nil {
			t.Error("expected error for empty filename")
		}
	})
}

func TestDecodeFileError(t *testing.T) {
	t.Parallel()

	t.Run("returns error for invalid yaml", func(t *testing.T) {
		t.Parallel()

		invalidYAML := strings.NewReader(":\n  invalid:\n  - [broken")

		var doc model.Document

		err := storage.DecodeFile(invalidYAML, &doc)
		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})
}

func TestListObjectsLocal(t *testing.T) {
	t.Parallel()

	t.Run("lists files sorted by modified date descending", func(t *testing.T) {
		t.Parallel()

		baseDir := t.TempDir()
		subDir := filepath.Join(baseDir, "articles")

		err := os.MkdirAll(subDir, 0750)
		if err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}

		for _, name := range []string{"alpha.yaml", "bravo.yaml"} {
			err := os.WriteFile(filepath.Join(subDir, name), []byte("title: "+name), 0600)
			if err != nil {
				t.Fatalf("failed to write %s: %v", name, err)
			}
		}

		older := time.Now().Add(-2 * time.Hour)
		newer := time.Now().Add(-1 * time.Hour)

		err = os.Chtimes(filepath.Join(subDir, "alpha.yaml"), older, older)
		if err != nil {
			t.Fatalf("failed to set mod time for alpha: %v", err)
		}

		err = os.Chtimes(filepath.Join(subDir, "bravo.yaml"), newer, newer)
		if err != nil {
			t.Fatalf("failed to set mod time for bravo: %v", err)
		}

		s := &storage.Storage{UseS3: false, BaseDir: baseDir}

		objects, err := s.ListObjects(context.Background(), "articles")
		if err != nil {
			t.Fatalf("ListObjects error: %v", err)
		}

		if len(objects) != 2 {
			t.Fatalf("expected 2 objects, got %d", len(objects))
		}

		if objects[0].LastModified.Before(*objects[1].LastModified) {
			t.Errorf("expected descending order: first=%v should be after second=%v",
				objects[0].LastModified, objects[1].LastModified)
		}
	})

	t.Run("skips subdirectories", func(t *testing.T) {
		t.Parallel()

		baseDir := t.TempDir()
		subDir := filepath.Join(baseDir, "projects")

		err := os.MkdirAll(filepath.Join(subDir, "nested"), 0750)
		if err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}

		err = os.WriteFile(filepath.Join(subDir, "only-file.yaml"), []byte("title: Test"), 0600)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		s := &storage.Storage{UseS3: false, BaseDir: baseDir}

		objects, err := s.ListObjects(context.Background(), "projects")
		if err != nil {
			t.Fatalf("ListObjects error: %v", err)
		}

		if len(objects) != 1 {
			t.Errorf("expected 1 object (skipping directory), got %d", len(objects))
		}
	})

	t.Run("creates directory if missing", func(t *testing.T) {
		t.Parallel()

		baseDir := t.TempDir()
		s := &storage.Storage{UseS3: false, BaseDir: baseDir}

		objects, err := s.ListObjects(context.Background(), "new-type")
		if err != nil {
			t.Fatalf("ListObjects error: %v", err)
		}

		if len(objects) != 0 {
			t.Errorf("expected 0 objects for newly created dir, got %d", len(objects))
		}

		info, statErr := os.Stat(filepath.Join(baseDir, "new-type"))
		if statErr != nil {
			t.Fatalf("expected directory to be created: %v", statErr)
		}

		if !info.IsDir() {
			t.Error("expected a directory, not a file")
		}
	})
}

func TestGetFileLocal(t *testing.T) {
	t.Parallel()

	t.Run("reads existing local file", func(t *testing.T) {
		t.Parallel()

		baseDir := t.TempDir()
		articlesDir := filepath.Join(baseDir, "articles")

		err := os.MkdirAll(articlesDir, 0750)
		if err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		err = os.WriteFile(filepath.Join(articlesDir, "test.yaml"), []byte("title: Hello"), 0600)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		s := &storage.Storage{UseS3: false, BaseDir: baseDir}

		file, err := s.GetFile(context.Background(), "articles/test.yaml")
		if err != nil {
			t.Fatalf("GetFile error: %v", err)
		}
		defer file.Close()

		content, err := os.ReadFile(file.Name())
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if string(content) != "title: Hello" {
			t.Errorf("expected 'title: Hello', got %q", string(content))
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		t.Parallel()

		s := &storage.Storage{UseS3: false, BaseDir: t.TempDir()}

		_, err := s.GetFile(context.Background(), "articles/missing.yaml")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
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

func TestDownloadS3FileLocalMode(t *testing.T) {
	t.Parallel()

	s := &storage.Storage{UseS3: false, BaseDir: t.TempDir()}

	err := s.DownloadS3File(context.Background(), "articles/test.yaml")
	if err != nil {
		t.Errorf("expected no error in local mode, got %v", err)
	}
}

func TestGetPreparedFile(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	articlesDir := filepath.Join(baseDir, "articles")

	err := os.MkdirAll(articlesDir, 0750)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	yamlContent := "title: Decoded Doc\nsubtitle: Sub\npreview: Preview text.\n"

	err = os.WriteFile(filepath.Join(articlesDir, "doc.yaml"), []byte(yamlContent), 0600)
	if err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	s := &storage.Storage{UseS3: false, BaseDir: baseDir}

	t.Run("GetPreparedFile decodes yaml", func(t *testing.T) {
		t.Parallel()

		var doc model.Document

		err := s.GetPreparedFile(context.Background(), "articles/doc.yaml", &doc)
		if err != nil {
			t.Fatalf("GetPreparedFile error: %v", err)
		}

		if doc.Title != "Decoded Doc" {
			t.Errorf("expected title 'Decoded Doc', got %q", doc.Title)
		}
	})
}

func TestGetDocumentBody(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	articlesDir := filepath.Join(baseDir, "articles")

	err := os.MkdirAll(articlesDir, 0750)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	mdContent := "# Test Article\n\nSome **bold** text."

	err = os.WriteFile(filepath.Join(articlesDir, "test.md"), []byte(mdContent), 0600)
	if err != nil {
		t.Fatalf("failed to write md file: %v", err)
	}

	s := &storage.Storage{UseS3: false, BaseDir: baseDir}

	t.Run("GetDocumentBodyRaw returns raw markdown", func(t *testing.T) {
		t.Parallel()

		raw, err := s.GetDocumentBodyRaw(context.Background(), "articles/test.yaml")
		if err != nil {
			t.Fatalf("GetDocumentBodyRaw error: %v", err)
		}

		if raw != mdContent {
			t.Errorf("expected raw markdown %q, got %q", mdContent, raw)
		}
	})

	t.Run("GetDocumentBody returns HTML", func(t *testing.T) {
		t.Parallel()

		html, err := s.GetDocumentBody(context.Background(), "articles/test.yaml")
		if err != nil {
			t.Fatalf("GetDocumentBody error: %v", err)
		}

		if !strings.Contains(html, "<strong>bold</strong>") {
			t.Errorf("expected HTML with <strong>, got %q", html)
		}

		if !strings.Contains(html, `class="category-title"`) {
			t.Errorf("expected styled h1, got %q", html)
		}
	})

	t.Run("GetDocumentBody returns error for missing file", func(t *testing.T) {
		t.Parallel()

		_, err := s.GetDocumentBody(context.Background(), "articles/nonexistent.yaml")
		if err == nil {
			t.Error("expected error for missing markdown file")
		}
	})
}

func getYAMLDocument() *fstest.MapFile {
	return &fstest.MapFile{
		Data: []byte(
			"title: Test Document\nsubtitle: Test Subtitle\n" +
				"preview: A brief preview.\ntags:\n  - test\n  - document\n",
		),
	}
}
