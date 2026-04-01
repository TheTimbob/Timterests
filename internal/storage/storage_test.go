package storage_test

import (
	"context"
	"os"
	"path/filepath"
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

	t.Run("decode YAML document", func(t *testing.T) {
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

		expectedBody := "Test Body"
		if document.Body != expectedBody {
			t.Errorf("Expected body '%s', got %v", expectedBody, document.Body)
		}
	})
	t.Run("write YAML document", func(t *testing.T) {
		t.Parallel()

		formData := map[string]any{
			"title":    "Test Document",
			"subtitle": "Test Subtitle",
			"body":     "Test Body",
			"tags":     []string{"test", "document"},
		}

		tempDir := t.TempDir()
		localFilePath := tempDir + "/test-document.yaml"

		err := storage.WriteYAMLDocument(localFilePath, formData)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify the file was created
		_, err = os.Stat(localFilePath)
		if os.IsNotExist(err) {
			t.Fatalf("File was not created: %v", err)
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
		Data: []byte(`title: Test Document
subtitle: Test Subtitle
body: Test Body
tags:
  - test
  - document
`),
	}
}
