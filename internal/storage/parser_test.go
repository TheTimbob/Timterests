package storage_test

import (
	"reflect"
	"strings"
	"testing"

	"timterests/internal/model"
	"timterests/internal/storage"
)

func getTestDocument() model.Document {
	return model.Document{
		Title:    "Test Document",
		Subtitle: "Test Subtitle",
		Body:     "This is a test body.\n\n- Item 1\n- Item 2\n\n## Subtitle\n\n[Link](http://example.com)",
		Tags:     []string{"test", "document"},
	}
}

func TestDocumentParser(t *testing.T) {
	t.Parallel()

	t.Run("body to HTML conversion", func(t *testing.T) {
		t.Parallel()

		// Use a local copy to avoid a race with the sibling sub-test that reads
		// the same document concurrently.
		document := getTestDocument()

		err := storage.BodyToHTML(&document)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedBody := `<p class="content-text">This is a test body.</p>
						<ul>
						<li class="content-text">Item 1</li>
						<li class="content-text">Item 2</li>
						</ul>
						<h2 class="category-subtitle">Subtitle</h2>
						<p class="content-text"><a href="http://example.com">Link</a></p>`

		// Normalize whitespace for comparison
		gotNormalized := strings.ReplaceAll(document.Body, " ", "")
		gotNormalized = strings.ReplaceAll(gotNormalized, "\n", "")
		gotNormalized = strings.ReplaceAll(gotNormalized, "\t", "")
		expectedNormalized := strings.ReplaceAll(expectedBody, " ", "")
		expectedNormalized = strings.ReplaceAll(expectedNormalized, "\n", "")
		expectedNormalized = strings.ReplaceAll(expectedNormalized, "\t", "")

		if gotNormalized != expectedNormalized {
			t.Errorf("BodyToHTML did not produce expected output.\nGot:\n%s\nExpected:\n%s", document.Body, expectedBody)
		}

		// Test incorrect type handling
		invalidInput := 42

		err = storage.BodyToHTML(&invalidInput)
		if err == nil {
			t.Fatalf("Expected error for invalid input type, got nil")
		}

		// Test empty body handling
		emptyDocument := model.Document{
			ID:       "",
			S3Key:    "",
			Title:    "Empty Document",
			Subtitle: "",
			Body:     "",
			Tags:     []string{},
		}

		err = storage.BodyToHTML(&emptyDocument)
		if err != nil {
			t.Fatalf("Expected no error for empty body, got %v", err)
		}

		if emptyDocument.Body != "" {
			t.Errorf("Expected empty body to remain empty, got %s", emptyDocument.Body)
		}
	})
	t.Run("get tags from document", func(t *testing.T) {
		t.Parallel()

		// Use a local copy to avoid a race with the sibling sub-test that writes
		// to the shared document concurrently.
		document := getTestDocument()

		var tags []string

		v := reflect.ValueOf(document)
		tags = storage.GetTags(v, tags)

		expectedTags := []string{"test", "document"}
		if len(tags) != len(expectedTags) {
			t.Fatalf("Expected %d tags, got %d", len(expectedTags), len(tags))
		}

		for i, tag := range expectedTags {
			if tags[i] != tag {
				t.Errorf("Expected tag %s at index %d, got %s", tag, i, tags[i])
			}
		}
	})
}

func TestHTMLParsing(t *testing.T) {
	t.Parallel()
	t.Run("remove HTML tags from string", func(t *testing.T) {
		t.Parallel()

		input := `<p>This is a <strong>test</strong> string with <a href="#">HTML</a> tags.</p>`
		expected := "This is a test string with HTML tags."

		result := storage.RemoveHTMLTags(input)

		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
	t.Run("filenames are sanitized and formatted", func(t *testing.T) {
		t.Parallel()

		longFilename := "This is a Very Long Filename without Special Characters"
		expectedLF := "this-is-a-very-long-filename-without-special-chara"

		sanitizedLF := storage.SanitizeFilename(longFilename)

		if sanitizedLF != expectedLF {
			t.Errorf("Expected sanitized filename to be %s, got %s", expectedLF, sanitizedLF)
		}

		specialCharFilename := "Inva|id/Filename!.yaml"
		expectedSCF := "filenameyaml"

		sanitizedSCF := storage.SanitizeFilename(specialCharFilename)

		if sanitizedSCF != expectedSCF {
			t.Errorf("Expected sanitized filename to be %s, got %s", expectedSCF, sanitizedSCF)
		}

		exploitFilename := "../../etc/.ssh/rsa_key"
		expectedEF := "rsa_key"

		sanitizedEF := storage.SanitizeFilename(exploitFilename)

		if sanitizedEF != expectedEF {
			t.Errorf("Expected sanitized filename to be %s, got %s", expectedEF, sanitizedEF)
		}

		missingFilename := ""
		expectedMF := ""

		sanitizedMF := storage.SanitizeFilename(missingFilename)

		if sanitizedMF == expectedMF {
			t.Errorf("Expected sanitized filename to not be empty")
		}
	})
}
