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
		Tags:     []string{"test", "document"},
	}
}

func TestMarkdownToHTML(t *testing.T) {
	t.Parallel()

	t.Run("markdown to HTML conversion", func(t *testing.T) {
		t.Parallel()

		input := []byte("This is a test body.\n\n- Item 1\n- Item 2\n\n## Subtitle\n\n[Link](http://example.com)")

		html, err := storage.MarkdownToHTML(input)
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
		gotNormalized := strings.ReplaceAll(html, " ", "")
		gotNormalized = strings.ReplaceAll(gotNormalized, "\n", "")
		gotNormalized = strings.ReplaceAll(gotNormalized, "\t", "")
		expectedNormalized := strings.ReplaceAll(expectedBody, " ", "")
		expectedNormalized = strings.ReplaceAll(expectedNormalized, "\n", "")
		expectedNormalized = strings.ReplaceAll(expectedNormalized, "\t", "")

		if gotNormalized != expectedNormalized {
			t.Errorf("MarkdownToHTML did not produce expected output.\nGot:\n%s\nExpected:\n%s", html, expectedBody)
		}
	})

	t.Run("empty body remains empty", func(t *testing.T) {
		t.Parallel()

		input := []byte("")

		html, err := storage.MarkdownToHTML(input)
		if err != nil {
			t.Fatalf("Expected no error for empty input, got %v", err)
		}

		if html != "" {
			t.Errorf("Expected empty HTML, got %s", html)
		}
	})
}

func TestGetTags(t *testing.T) {
	t.Parallel()

	t.Run("get tags from document", func(t *testing.T) {
		t.Parallel()

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

	t.Run("get tags from embedded document struct", func(t *testing.T) {
		t.Parallel()

		article := model.Article{
			Document: model.Document{
				Tags: []string{"go", "web"},
			},
			Date: "2026-01-01",
		}

		var tags []string

		v := reflect.ValueOf(article)
		tags = storage.GetTags(v, tags)

		if len(tags) != 2 || tags[0] != "go" || tags[1] != "web" {
			t.Errorf("Expected [go web], got %v", tags)
		}
	})

	t.Run("deduplicates tags", func(t *testing.T) {
		t.Parallel()

		document := model.Document{
			Tags: []string{"go", "web"},
		}

		existing := []string{"go"}

		v := reflect.ValueOf(document)
		tags := storage.GetTags(v, existing)

		if len(tags) != 2 {
			t.Errorf("Expected 2 unique tags, got %d: %v", len(tags), tags)
		}
	})

	t.Run("returns existing tags for struct without tags field", func(t *testing.T) {
		t.Parallel()

		type NoTags struct {
			Name string
		}

		v := reflect.ValueOf(NoTags{Name: "test"})
		existing := []string{"existing"}
		tags := storage.GetTags(v, existing)

		if len(tags) != 1 || tags[0] != "existing" {
			t.Errorf("Expected [existing], got %v", tags)
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
