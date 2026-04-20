// Package ai provides AI-powered content generation and suggestion functionality.
package ai

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// LoadAPIKey loads the OpenAI API key from environment variables.
func LoadAPIKey() (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load env file: %w", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY not found in environment variables")
	}

	return apiKey, nil
}

// GenerateSuggestion generates content suggestions using OpenAI's API.
// systemInstruction is the system prompt content (not a filename).
func GenerateSuggestion(ctx context.Context, prompt, systemInstruction string) (string, error) {
	apiKey, envLoadErr := LoadAPIKey()
	if envLoadErr != nil {
		return "", envLoadErr
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemInstruction),
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return "", errors.New("no choices returned from OpenAI API")
	}

	return CleanSuggestion(chatCompletion.Choices[0].Message.Content), nil
}

// CleanSuggestion strips markdown and YAML formatting from AI-generated text
// so it renders as plain prose in the suggestion display.
func CleanSuggestion(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))

	inFrontmatter := false
	inCodeBlock := false
	frontmatterDone := false

	for i, line := range lines {
		// Strip YAML frontmatter (--- ... ---) at the top of the document.
		if !frontmatterDone && i == 0 && strings.TrimSpace(line) == "---" {
			inFrontmatter = true

			continue
		}

		if inFrontmatter {
			if strings.TrimSpace(line) == "---" {
				inFrontmatter = false
				frontmatterDone = true
			}

			continue
		}

		// Remove fenced code block delimiters; keep the content as plain text.
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock

			continue
		}

		if inCodeBlock {
			out = append(out, line)

			continue
		}

		// Strip ATX headings (# Heading).
		stripped := strings.TrimLeft(line, "#")
		if len(stripped) < len(line) {
			line = strings.TrimSpace(stripped)
		}

		// Strip bold/italic markers.
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "__", "")
		line = strings.ReplaceAll(line, "*", "")

		// Strip inline backticks.
		line = strings.ReplaceAll(line, "`", "")

		out = append(out, line)
	}

	return strings.TrimSpace(strings.Join(out, "\n"))
}

// GetInstruction reads and returns the content of a prompt instruction file.
// file must be a plain filename with no path components.
func GetInstruction(file string) (string, error) {
	file = filepath.Base(filepath.Clean(file))
	if strings.TrimSpace(file) == "" || strings.Contains(file, string(filepath.Separator)) {
		return "", fmt.Errorf("invalid prompt file: %q", file)
	}

	promptsFS := os.DirFS("prompts")

	content, err := fs.ReadFile(promptsFS, file)
	if err != nil {
		return "", fmt.Errorf("failed to read instruction file: %w", err)
	}

	return string(content), nil
}

// GetInstructionOptionList retrieves a list of available prompt files and their display titles.
func GetInstructionOptionList(promptPath string) ([]string, []string, error) {
	entries, err := os.ReadDir(promptPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read prompt directory: %w", err)
	}

	var (
		filePaths []string
		titles    []string
	)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Only include .txt files.
		if filepath.Ext(name) == ".txt" {
			titleName := FormatPromptFileName(name)
			titles = append(titles, titleName)

			filePaths = append(filePaths, name)
		}
	}

	return titles, filePaths, nil
}

// FormatPromptFileName converts a prompt filename to a human-readable title.
func FormatPromptFileName(promptFile string) string {
	// Extract the base name without extension
	name := filepath.Base(promptFile)
	name = strings.TrimSuffix(name, filepath.Ext(name))

	// Replace underscores with spaces and convert to title case
	name = strings.ReplaceAll(name, "_", " ")
	caser := cases.Title(language.English)
	name = caser.String(name)

	return name
}
