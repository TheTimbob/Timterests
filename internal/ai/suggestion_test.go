package ai

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("OPENAI_API_KEY=test-key\n"), 0600); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}

	oldWd, _ := os.Getwd()
	err := os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer func() {
		err := os.Chdir(oldWd)
		if err != nil {
			t.Fatalf("Chdir failed: %v", err)
		}
	}()

	apiKey, err := LoadAPIKey()
	if err != nil {
		t.Fatalf("LoadAPIKey error: %v", err)
	}
	if apiKey != "test-key" {
		t.Fatalf("got %q, want %q", apiKey, "test-key")
	}
}

func TestGetInstruction(t *testing.T) {
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to ensure prompts dir: %v", err)
	}

	oldWd, _ := os.Getwd()
	err := os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer func() {
		err := os.Chdir(oldWd)
		if err != nil {
			t.Fatalf("Chdir failed: %v", err)
		}
	}()

	tmp, err := os.CreateTemp(promptsDir, "instruction-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	defer func(name string) {
		if err := os.Remove(name); err != nil {
			t.Fatalf("failed to remove temp file %s: %v", name, err)
		}
	}(tmp.Name())

	content := "System instruction line"
	if _, err := tmp.WriteString(content); err != nil {
		t.Fatalf("write to tmp failed: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	instruction, err := GetInstruction(filepath.Base(tmp.Name()))
	if err != nil {
		t.Fatalf("Failed to get instruction: %v", err)
	}
	if instruction == "" {
		t.Fatal("Instruction content is empty")
	}
}

func TestGetInstructionOptionList(t *testing.T) {
	// Determine project root (this test file lives in <root>/internal/ai)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	internalDir := filepath.Dir(wd)
	rootDir := filepath.Dir(internalDir)
	promptsDir := filepath.Join(rootDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to ensure prompts dir: %v", err)
	}

	tmp, err := os.CreateTemp(promptsDir, "instruction-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	tmpName := tmp.Name()
	_ = tmp.Close()
	defer func() {
		if err := os.Remove(tmpName); err != nil {
			t.Fatalf("failed to remove temp file %s: %v", tmpName, err)
		}
	}()

	titles, filePaths, err := GetInstructionOptionList(promptsDir)
	if err != nil {
		t.Fatalf("Failed to get instruction option lists: %v", err)
	}
	if len(titles) == 0 || len(filePaths) == 0 {
		t.Fatal("Instruction option lists are empty")
	}
	title := formatPromptFileName(tmpName)
	if !slices.Contains(titles, title) {
		t.Fatalf("Expected %s in titles: %v", title, titles)
	}

	if !slices.Contains(filePaths, filepath.Base(tmpName)) {
		t.Fatalf("Expected %s in filePaths: %v", filepath.Base(tmpName), filePaths)
	}
}

func TestFormatPromptFileName(t *testing.T) {
	promptFile := "best_article.txt"
	expected := "Best Article"
	formatted := formatPromptFileName(promptFile)

	if formatted != expected {
		t.Fatalf("got %q, want %q", formatted, expected)
	}
}
