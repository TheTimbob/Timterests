package ai_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"timterests/internal/ai"
)

func TestCleanSuggestion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text passthrough",
			input: "This is a plain sentence.",
			want:  "This is a plain sentence.",
		},
		{
			name:  "strips YAML frontmatter",
			input: "---\ntitle: Test\n---\nActual content here.",
			want:  "Actual content here.",
		},
		{
			name:  "strips ATX headings",
			input: "# Big Heading\n## Sub\nBody text.",
			want:  "Big Heading\nSub\nBody text.",
		},
		{
			name:  "strips bold and italic markers",
			input: "This is **bold** and *italic* and __also bold__.",
			want:  "This is bold and italic and also bold.",
		},
		{
			name:  "strips inline backticks",
			input: "Call `fmt.Println` to print.",
			want:  "Call fmt.Println to print.",
		},
		{
			name:  "removes code fence delimiters, keeps content",
			input: "Before.\n```go\nfmt.Println(\"hi\")\n```\nAfter.",
			want:  "Before.\nfmt.Println(\"hi\")\nAfter.",
		},
		{
			name:  "frontmatter only at document start",
			input: "Some text.\n---\nnot: frontmatter\n---\nMore text.",
			want:  "Some text.\n---\nnot: frontmatter\n---\nMore text.",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "trims surrounding whitespace",
			input: "\n\nHello world.\n\n",
			want:  "Hello world.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ai.CleanSuggestion(tc.input)
			if got != tc.want {
				t.Fatalf("CleanSuggestion(%q)\ngot:  %q\nwant: %q", tc.input, got, tc.want)
			}
		})
	}
}

// Not parallel: changes working directory.
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
	// Not parallel: changes working directory.
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
