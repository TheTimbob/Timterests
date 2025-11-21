package ai

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func LoadAPIKey() (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY not found in environment variables")
	}
	return apiKey, nil
}

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
		return "", err
	}

	if len(chatCompletion.Choices) == 0 {
		return "", errors.New("no choices returned from OpenAI API")
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

func GetInstruction(file string) (string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GetInstructionOptionList(promptPath string) ([]string, error) {
	entries, err := os.ReadDir(promptPath)
	if err != nil {
		return nil, err
	}

	var options []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Only include .txt files.
		if filepath.Ext(name) == ".txt" {
			fileWithoutExt := strings.TrimSuffix(name, ".txt")
			options = append(options, fileWithoutExt)
		}
	}
	return options, nil
}
