package storage

import (
	"reflect"
	"strings"
	"testing"

	"timterests/internal/types"
)

func getTestDocument() types.Document {
	return types.Document{
		Title:    "Test Document",
		Subtitle: "Test Subtitle",
		Body:     "This is a test body.\n\n- Item 1\n- Item 2\n\n## Subtitle\n\n[Link](http://example.com)",
		Tags:     []string{"test", "document"},
	}
}

func TestBodyToHTML(t *testing.T) {
	document := getTestDocument()

	err := BodyToHTML(&document)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedBody := `<p class="content-text">This is a test body.</p>
					<ul>
					<li class="content-text">Item 1</li>
					<li class="content-text">Item 2</li>
					</ul>
					<h2 class="category-subtitle">Subtitle</h2>
					<p class="content-text"><a class="hyperlink" href="http://example.com">Link</a></p>`

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
	err = BodyToHTML(&invalidInput)
	if err == nil {
		t.Fatalf("Expected error for invalid input type, got nil")
	}

	// Test empty body handling
	emptyDocument := types.Document{
		Title: "Empty Document",
		Body:  "",
	}
	err = BodyToHTML(&emptyDocument)
	if err != nil {
		t.Fatalf("Expected no error for empty body, got %v", err)
	}
	if emptyDocument.Body != "" {
		t.Errorf("Expected empty body to remain empty, got %s", emptyDocument.Body)
	}
}

func TestGetTags(t *testing.T) {
	document := getTestDocument()
	var tags []string
	v := reflect.ValueOf(document)
	tags = GetTags(v, tags)

	expectedTags := []string{"test", "document"}
	if len(tags) != len(expectedTags) {
		t.Fatalf("Expected %d tags, got %d", len(expectedTags), len(tags))
	}

	for i, tag := range expectedTags {
		if tags[i] != tag {
			t.Errorf("Expected tag %s at index %d, got %s", tag, i, tags[i])
		}
	}
}

func TestRemoveHTMLTags(t *testing.T) {
	input := `<p>This is a <strong>test</strong> string with <a href="#">HTML</a> tags.</p>`
	expected := "This is a test string with HTML tags."
	result := RemoveHTMLTags(input)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestSanitizeFilename(t *testing.T) {
	longFilename := "This is a Very Long Filename without Special Characters"
	expectedLF := "this-is-a-very-long-filename-without-special-chara"
	sanitizedLF := SanitizeFilename(longFilename)
	if sanitizedLF != expectedLF {
		t.Errorf("Expected sanitized filename to be %s, got %s", expectedLF, sanitizedLF)
	}

	specialCharFilename := "Inva|id/Filename!.yaml"
	expectedSCF := "filenameyaml"
	sanitizedSCF := SanitizeFilename(specialCharFilename)
	if sanitizedSCF != expectedSCF {
		t.Errorf("Expected sanitized filename to be %s, got %s", expectedSCF, sanitizedSCF)
	}

	exploitFilename := "../../etc/.ssh/rsa_key"
	expectedEF := "rsa_key"
	sanitizedEF := SanitizeFilename(exploitFilename)
	if sanitizedEF != expectedEF {
		t.Errorf("Expected sanitized filename to be %s, got %s", expectedEF, sanitizedEF)
	}

	missingFilename := ""
	expectedMF := ""
	sanitizedMF := SanitizeFilename(missingFilename)
	if sanitizedMF == expectedMF {
		t.Errorf("Expected sanitized filename to not be empty")
	}
}
