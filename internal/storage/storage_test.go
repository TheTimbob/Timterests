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

	promptsDir := t.TempDir()

	// Write a prompt file for "articles"
	promptContent := "You are a helpful writing assistant for articles."

	err := os.WriteFile(filepath.Join(promptsDir, "articles.txt"), []byte(promptContent), 0600)
	if err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	s := &storage.Storage{
		UseS3:      false,
		PromptsDir: promptsDir,
	}

	t.Run("returns prompt content for valid doc type", func(t *testing.T) {
		t.Parallel()

		got, err := s.GetPromptContent(context.Background(), "articles")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != promptContent {
			t.Errorf("got %q, want %q", got, promptContent)
		}
	})

	t.Run("returns error for unsupported doc type", func(t *testing.T) {
		t.Parallel()

		_, err := s.GetPromptContent(context.Background(), "unknown-type")
		if err == nil {
			t.Fatal("expected error for unsupported doc type, got nil")
		}
	})

	t.Run("returns error when prompt file is missing", func(t *testing.T) {
		t.Parallel()

		emptyDir := t.TempDir()
		sNoFile := &storage.Storage{
			UseS3:      false,
			PromptsDir: emptyDir,
		}

		_, err := sNoFile.GetPromptContent(context.Background(), "projects")
		if err == nil {
			t.Fatal("expected error when prompt file is missing, got nil")
		}
	})
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
