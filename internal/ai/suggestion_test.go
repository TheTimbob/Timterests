package ai_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"timterests/internal/ai"
)

//nolint:paralleltest // changing working directory
func TestLoadAPIKey(t *testing.T) {
	t.Run("load API key from .env file", func(t *testing.T) {
		tmpDir := t.TempDir()

		envPath := filepath.Join(tmpDir, ".env")

		err := os.WriteFile(envPath, []byte("OPENAI_API_KEY=test-key\n"), 0600)
		if err != nil {
			t.Fatalf("failed to create .env: %v", err)
		}

		t.Chdir(tmpDir)

		apiKey, err := ai.LoadAPIKey()
		if err != nil {
			t.Fatalf("LoadAPIKey error: %v", err)
		}

		if apiKey != "test-key" {
			t.Fatalf("got %q, want %q", apiKey, "test-key")
		}
	})
}

func TestPromptOperations(t *testing.T) {
	// This test changes working directory in subtests and therefore
	// must not run in parallel with other tests.

	//nolint:paralleltest // changing working directory
	t.Run("get instruction from file", func(t *testing.T) {
		tmpDir := t.TempDir()

		promptsDir := filepath.Join(tmpDir, "prompts")

		err := os.MkdirAll(promptsDir, 0750)
		if err != nil {
			t.Fatalf("failed to ensure prompts dir: %v", err)
		}

		t.Chdir(tmpDir)

		tmp, err := os.CreateTemp(promptsDir, "instruction-*.txt")
		if err != nil {
			t.Fatalf("CreateTemp failed: %v", err)
		}
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				t.Fatalf("failed to remove temp file %s: %v", name, err)
			}
		}(tmp.Name())

		content := "System instruction line"

		_, err = tmp.WriteString(content)
		if err != nil {
			t.Fatalf("write to tmp failed: %v", err)
		}

		err = tmp.Close()
		if err != nil {
			t.Fatalf("failed to close temp file: %v", err)
		}

		instruction, err := ai.GetInstruction(filepath.Base(tmp.Name()))
		if err != nil {
			t.Fatalf("Failed to get instruction: %v", err)
		}

		if instruction == "" {
			t.Fatal("Instruction content is empty")
		}
	})

	t.Run("list instruction options", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		promptsDir := filepath.Join(tmpDir, "prompts")

		err := os.MkdirAll(promptsDir, 0750)
		if err != nil {
			t.Fatalf("failed to ensure prompts dir: %v", err)
		}

		tmp, err := os.CreateTemp(promptsDir, "instruction-*.txt")
		if err != nil {
			t.Fatalf("CreateTemp failed: %v", err)
		}

		tmpName := tmp.Name()
		_ = tmp.Close()

		defer func() {
			err := os.Remove(tmpName)
			if err != nil {
				t.Fatalf("failed to remove temp file %s: %v", tmpName, err)
			}
		}()

		titles, filePaths, err := ai.GetInstructionOptionList(promptsDir)
		if err != nil {
			t.Fatalf("Failed to get instruction option lists: %v", err)
		}

		if len(titles) == 0 || len(filePaths) == 0 {
			t.Fatal("Instruction option lists are empty")
		}

		title := ai.FormatPromptFileName(tmpName)
		if !slices.Contains(titles, title) {
			t.Fatalf("Expected %s in titles: %v", title, titles)
		}

		if !slices.Contains(filePaths, filepath.Base(tmpName)) {
			t.Fatalf("Expected %s in filePaths: %v", filepath.Base(tmpName), filePaths)
		}
	})

	t.Run("format prompt filename", func(t *testing.T) {
		t.Parallel()

		promptFile := "best_article.txt"
		expected := "Best Article"
		formatted := ai.FormatPromptFileName(promptFile)

		if formatted != expected {
			t.Fatalf("got %q, want %q", formatted, expected)
		}
	})
}
