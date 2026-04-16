package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

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

		// Body is not stored in YAML — it should be empty after decode
		if doc.Body != "" {
			t.Errorf("Expected empty body after YAML decode, got %q", doc.Body)
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

func getYAMLDocument() *fstest.MapFile {
	return &fstest.MapFile{
		Data: []byte(
			"title: Test Document\nsubtitle: Test Subtitle\n" +
				"preview: A brief preview.\ntags:\n  - test\n  - document\n",
		),
	}
}
