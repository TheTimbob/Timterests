package web

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"timterests/internal/auth"
	"timterests/internal/storage"
)

// DownloadDocumentHandler handles document download requests for authenticated users.
func DownloadDocumentHandler(w http.ResponseWriter, r *http.Request, title string, a *auth.Auth) {
	// Only admins can download documents
	if !a.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	fileName := storage.SanitizeFilename(title) + ".md"
	filePath := filepath.Join("storage", fileName)

	// Set headers to force download
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "text/markdown")

	http.ServeFile(w, r, filePath)
}

// DownloadNewDocumentHandler handles requests to download a new document based on form data.
func DownloadNewDocumentHandler(w http.ResponseWriter, r *http.Request, a *auth.Auth) {
	// Only admins can download documents
	if !a.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)

		return
	}

	filename := storage.SanitizeFilename("") + ".md"

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

	localFilePath := filepath.Join("storage", filename)

	err = storage.WriteMarkdownDocument(localFilePath, formData)
	if err != nil {
		http.Error(w, "Failed to write Markdown document", http.StatusInternalServerError)

		return
	}

	// Cleanup temporary file after serving
	defer func() {
		err := os.Remove(localFilePath)
		if err != nil {
			log.Printf("Failed to remove temporary file: %v", err)
		}
	}()

	title, ok := formData["title"].(string)
	if !ok || title == "" {
		http.Error(w, "Missing or invalid title field", http.StatusBadRequest)

		return
	}

	downloadFilename := storage.SanitizeFilename(title) + ".md"

	// Set headers to force download
	w.Header().Set("Content-Disposition", "attachment; filename=\""+downloadFilename+"\"")
	w.Header().Set("Content-Type", "text/markdown")

	http.ServeFile(w, r, localFilePath)
}
