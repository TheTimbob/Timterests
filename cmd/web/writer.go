package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"timterests/internal/storage"

	"github.com/a-h/templ"
)

const dateFormat = "01-02-2006"

func WriterPageHandler(w http.ResponseWriter, r *http.Request, docType string) {
	var component templ.Component

	// TODO - Uncomment this before release
	// if !IsAuthenticated(r) {
	// 	http.Redirect(w, r, "/login", http.StatusSeeOther)
	// 	return
	// }

	if r.Header.Get("HX-Request") == "true" {
		component = FormContentByType(docType)
	} else {
		component = WriterPage()
	}

	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in WriterPage: %v", err)
	}
}

func WriteDocumentHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		log.Printf("Error parsing form: %v", err)
		return
	}

	formData := make(map[string]any)
	for key, values := range r.Form {
		if key == "tags" {
			tags := strings.Split(values[0], ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
			formData[key] = tags
		} else if len(values) > 0 {
			formData[key] = values[0]
		}
	}

	title, ok := formData["title"].(string)
	if !ok || strings.TrimSpace(title) == "" {
		http.Error(w, "Invalid or missing title in form data", http.StatusBadRequest)
		log.Printf("Invalid or missing title in form data")
		return
	}

	timestamp := time.Now().Format(dateFormat)
	sanitizedTitle := storage.SanitizeFilename(title)
	objectKey := fmt.Sprintf("%s-%s.yaml", sanitizedTitle, timestamp)

	err := storage.WriteYAMLDocument(storageInstance, objectKey, formData)
	if err != nil {
		http.Error(w, "Failed to save document", http.StatusInternalServerError)
		log.Printf("Error writing document: %v", err)
		return
	}
	http.Redirect(w, r, "/writer", http.StatusSeeOther)
}
