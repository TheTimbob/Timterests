package storage_test

import (
	"os"
	"testing"
	"testing/fstest"

	"timterests/internal/model"
	"timterests/internal/storage"
)

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
func TestDecodeFile(t *testing.T) {
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
}

func TestWriteYAMLDocument(t *testing.T) {
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
}
