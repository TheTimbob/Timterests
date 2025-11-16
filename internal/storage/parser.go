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

// Function to convert body text to HTML
func BodyToHTML(document any) error {
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)

	// Get reflect value of the document
	v := reflect.ValueOf(document)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("document must be a non-nil pointer to a struct")
	}
	v = v.Elem()

	// Check if the underlying value is a struct before getting Body
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("document must point to a struct")
	}

	// Set the Body field to the modified HTML content
	bodyField := v.FieldByName("Body")
	if !bodyField.IsValid() || bodyField.Kind() != reflect.String {
		return fmt.Errorf("document does not have a valid Body field of type string")
	}

	body := bodyField.String()
	if err := md.Convert([]byte(body), &buf); err != nil {
		log.Printf("failed to convert body to HTML: %v", err)
		return fmt.Errorf("conversion error: %w", err)
	}

	body = buf.String()

	body = strings.ReplaceAll(body, "<p>", `<p class="content-text">`)
	body = strings.ReplaceAll(body, "<h2>", `<h2 class="category-subtitle">`)
	body = strings.ReplaceAll(body, "<a ", `<a class="hyperlink"`)
	body = strings.ReplaceAll(body, "<li>", `<li class="content-text">- `)

	if bodyField.CanSet() {
		bodyField.SetString(body)
	} else {
		return fmt.Errorf("body field cannot be set in the document struct")
	}

	return nil
}

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
	for i := 0; i < field.Len(); i++ {
		tag := field.Index(i).String()
		if !slices.Contains(tags, tag) {
			tags = append(tags, tag)
		}
	}

	return tags
}

func RemoveHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

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
