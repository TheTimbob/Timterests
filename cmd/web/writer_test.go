package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"timterests/cmd/web"
)

func TestWriterSuggestionWithMultipleDocTypes(t *testing.T) {
	// Prevent the test from hitting the real OpenAI API if a key is set in the environment or .env file.
	t.Setenv("OPENAI_API_KEY", "")

	ctx := context.Background()
	s := testSetup(t, ctx)
	a, addCookie := testAuthentication(t)

	// Create a temp prompts directory with a distinct file for each doc type.
	// This verifies the handler maps each docType to its correct <docType>.txt file.
	promptsDir := t.TempDir()
	for _, docType := range []string{"articles", "projects", "reading-list", "letters"} {
		promptFile := filepath.Join(promptsDir, docType+".txt")
		if err := os.WriteFile(promptFile, []byte("test system prompt for "+docType), 0600); err != nil {
			t.Fatalf("failed to write prompt file for %s: %v", docType, err)
		}
	}
	s.PromptsDir = promptsDir

	tests := []struct {
		docType            string
		expectedPromptFile string
	}{
		{docType: "articles", expectedPromptFile: "articles.txt"},
		{docType: "projects", expectedPromptFile: "projects.txt"},
		{docType: "reading-list", expectedPromptFile: "reading-list.txt"},
		{docType: "letters", expectedPromptFile: "letters.txt"},
	}

	for _, tc := range tests {
		t.Run(tc.docType+" uses "+tc.expectedPromptFile, func(t *testing.T) {
			t.Parallel()

			form := url.Values{
				"document-type": {tc.docType},
				"body":          {"body content for AI suggestion test"},
			}
			req := httptest.NewRequest(http.MethodPost, "/writer/suggest", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			addCookie(req)
			rec := httptest.NewRecorder()

			web.WriterSuggestionHandler(rec, req, *s, a)

			// GetPromptContent succeeds when the correct prompt file is found.
			// A "temporarily unavailable" response indicates the prompt was NOT loaded,
			// meaning the docType→filename mapping failed.
			if strings.Contains(rec.Body.String(), "AI suggestions are temporarily unavailable") {
				t.Errorf("docType %q: expected prompt file %q to be loaded, but got 'temporarily unavailable'",
					tc.docType, tc.expectedPromptFile)
			}
		})
	}
}
