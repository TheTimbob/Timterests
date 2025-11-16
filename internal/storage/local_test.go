package storage

import (
	"os"
	"testing"
	"testing/fstest"
	"timterests/internal/types"
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

	var document types.Document
	var filename = "document.yaml"

	fs := &fstest.MapFS{
		filename: getYAMLDocument(),
	}

	file, err := fs.Open(filename)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = DecodeFile(file, &document)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedBody := "Test Body"
	if document.Body != expectedBody {
		t.Errorf("Expected body '%s', got %v", expectedBody, document.Body)
	}
}

func TestWriteYAMLDocument(t *testing.T) {
	formData := map[string]any{
		"title":    "Test Document",
		"subtitle": "Test Subtitle",
		"body":     "Test Body",
		"tags":     []string{"test", "document"},
	}

	tempDir := t.TempDir()
	localFilePath := tempDir + "/test-document.yaml"

	err := WriteYAMLDocument(localFilePath, formData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
		t.Fatalf("File was not created: %v", err)
	}
}
