package storage

import (
	"bytes"
	"fmt"
	"log"
	"path"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

// MarkdownToHTML converts raw markdown bytes to styled HTML.
func MarkdownToHTML(content []byte) (string, error) {
	var buf bytes.Buffer

	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)

	err := md.Convert(content, &buf)
	if err != nil {
		log.Printf("failed to convert markdown to HTML: %v", err)

		return "", fmt.Errorf("conversion error: %w", err)
	}

	body := buf.String()
	body = strings.ReplaceAll(body, "<p>", `<p class="content-text">`)
	body = strings.ReplaceAll(body, "<h1>", `<h1 class="category-title">`)
	body = strings.ReplaceAll(body, "<h2>", `<h2 class="category-subtitle">`)
	body = strings.ReplaceAll(body, "<li>", `<li class="content-text">`)

	return body, nil
}

// GetTags extracts tags from a struct value, checking embedded Document if needed.
func GetTags(v reflect.Value, tags []string) []string {
	field := v.FieldByName("Tags")

	// If the Tags field is not directly on the struct, check the embedded Document
	if !field.IsValid() {
		embeddedDoc := v.FieldByName("Document")
		if embeddedDoc.IsValid() {
			field = embeddedDoc.FieldByName("Tags")
		} else {
			return tags
		}
	}

	// Create a list of tags
	for i := range field.Len() {
		tag := field.Index(i).String()
		if !slices.Contains(tags, tag) {
			tags = append(tags, tag)
		}
	}

	return tags
}

// RemoveHTMLTags strips all HTML tags from a string.
func RemoveHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)

	return re.ReplaceAllString(s, "")
}

// SanitizeFilename sanitizes a filename for safe file system use.
func SanitizeFilename(filename string) string {
	// Strip any directory components from the filename to prevent directory traversal attacks.
	// This ensures only the base filename is sanitized.
	filename = path.Base(filename)
	filename = strings.ToLower(filename)
	filename = strings.ReplaceAll(filename, " ", "-")

	reg := regexp.MustCompile("[^a-z0-9-_]")
	filename = reg.ReplaceAllString(filename, "")

	const maxLength = 50
	if len(filename) > maxLength {
		filename = filename[:maxLength]
	}

	filename = strings.Trim(filename, ".-")

	// Ensure filename is not empty after trimming
	if filename == "" {
		return "unnamed-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	return filename
}
