package web_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"timterests/internal/storage"
)

func TestGetPromptContent(t *testing.T) {
	ctx := context.Background()

	// Create a temporary directory for prompts
	promptsDir := t.TempDir()

	// Create prompt files for each document type
	docTypes := []string{"articles", "projects", "reading-list", "letters"}
	for _, docType := range docTypes {
		promptFile := filepath.Join(promptsDir, docType+".txt")

		content := "test prompt for " + docType

		err := os.WriteFile(promptFile, []byte(content), 0600)
		if err != nil {
			t.Fatalf("failed to write prompt file for %s: %v", docType, err)
		}
	}

	// Create a storage instance with the temp prompts directory
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

			if !strings.Contains(content, tc.expectContent) {
				t.Errorf("expected content to contain %q, got %q", tc.expectContent, content)
			}
		})
	}
}
