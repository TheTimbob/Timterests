package web

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"timterests/internal/auth"
	"timterests/internal/storage"
)

func DownloadDocumentHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Only admins can download documents
	if !auth.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fileName := storage.SanitizeFilename(title) + ".yaml"
	filePath := filepath.Join("s3", fileName)

	// Set headers to force download
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "application/x-yaml")

	http.ServeFile(w, r, filePath)
}

func DownloadNewDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Only admins can download documents
	if !auth.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	filename := storage.SanitizeFilename("") + ".yaml"

	// Convert url.Values to map[string]any
	formData := make(map[string]any)
	for key, values := range r.Form {
		if len(values) == 1 {
			formData[key] = values[0]
		} else {
			formData[key] = values
		}
	}
	delete(formData, "document-type")

	localFilePath := path.Join("s3", filename)
	err := storage.WriteYAMLDocument(localFilePath, formData)
	if err != nil {
		http.Error(w, "Failed to write YAML document", http.StatusInternalServerError)
		return
	}

	// Cleanup temporary file after serving
	defer func() {
		if err := os.Remove(localFilePath); err != nil {
			fmt.Printf("Failed to remove temporary file: %v", err)
		}
	}()

	title, ok := formData["title"].(string)
	if !ok || title == "" {
		http.Error(w, "Missing or invalid title field", http.StatusBadRequest)
		return
	}
	downloadFilename := storage.SanitizeFilename(title) + ".yaml"

	// Set headers to force download
	w.Header().Set("Content-Disposition", "attachment; filename=\""+downloadFilename+"\"")
	w.Header().Set("Content-Type", "application/x-yaml")

	http.ServeFile(w, r, localFilePath)
}
