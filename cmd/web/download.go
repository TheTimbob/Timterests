package web

import (
	"fmt"
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
// It writes only the Markdown body (with title/subtitle headers) to a temporary file and serves it.
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

	title := r.FormValue("title")
	subtitle := r.FormValue("subtitle")
	body := r.FormValue("body")

	if title == "" {
		http.Error(w, "Missing or invalid title field", http.StatusBadRequest)

		return
	}

	downloadFilename := storage.SanitizeFilename(title) + ".md"

	f, err := os.CreateTemp("", "download-*.md")
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		log.Printf("DownloadNewDocumentHandler: failed to create temp file: %v", err)

		return
	}

	// Cleanup temporary file after serving
	defer func() {
		removeErr := os.Remove(f.Name())
		if removeErr != nil {
			log.Printf("Failed to remove temporary file: %v", removeErr)
		}
	}()

	_, err = fmt.Fprintf(f, "# %s\n## %s\n\n%s", title, subtitle, body)

	closeErr := f.Close()
	if closeErr != nil {
		log.Printf("DownloadNewDocumentHandler: failed to close temp file: %v", closeErr)
	}

	if err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		log.Printf("DownloadNewDocumentHandler: failed to write temp file: %v", err)

		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+downloadFilename+"\"")
	w.Header().Set("Content-Type", "text/markdown")

	http.ServeFile(w, r, f.Name())
}
