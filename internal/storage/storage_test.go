package storage_test

import (
	"os"
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
