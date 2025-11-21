package ai

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestLoadAPIKey(t *testing.T) {
	if err := os.WriteFile(".env", []byte("OPENAI_API_KEY=test-key\n"), 0600); err != nil {
		t.Fatalf("failed to create local .env: %v", err)
	}
	defer func() {
		if err := os.Remove(".env"); err != nil {
			t.Fatalf("failed to remove .env: %v", err)
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
	tmp, err := os.CreateTemp("", "instruction-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	defer func() {
		if err := os.Remove(tmp.Name()); err != nil {
			t.Fatalf("failed to remove temp file %s: %v", tmp.Name(), err)
		}
	}()
	content := "System instruction line"
	if _, err := tmp.WriteString(content); err != nil {
		t.Fatalf("write to tmp failed: %v", err)
	}
	_ = tmp.Close()

	instruction, err := GetInstruction(tmp.Name())
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

	options, err := GetInstructionOptionList(promptsDir)
	if err != nil {
		t.Fatalf("Failed to get instruction option list: %v", err)
	}
	if len(options) == 0 {
		t.Fatal("Instruction option list is empty")
	}
	fileName := strings.TrimSuffix(filepath.Base(tmpName), ".txt")
	if slices.Contains(options, fileName) == false {
		t.Fatalf("Expected %s in options: %v", fileName, options)
	}
}
