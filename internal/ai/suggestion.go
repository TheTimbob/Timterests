// Package ai provides AI-powered content generation and suggestion functionality.
package ai

import (
	"context"
	"errors"
	"fmt"
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
func GenerateSuggestion(ctx context.Context, prompt, instructionFile string) (string, error) {
	apiKey, envLoadErr := LoadAPIKey()
	if envLoadErr != nil {
		return "", envLoadErr
	}

	systemInstruction, err := GetInstruction(instructionFile)
	if err != nil {
		return "", err
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

	return chatCompletion.Choices[0].Message.Content, nil
}

// GetInstruction reads and returns the content of a prompt instruction file.
func GetInstruction(file string) (string, error) {
	// Ensure only filename, no path components
	file = filepath.Base(filepath.Clean(file))
	file = filepath.Join("prompts", file)

	content, err := os.ReadFile(file) // #nosec G304 -- file path is cleaned, reduced to basename, and restricted to prompts/ directory
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
