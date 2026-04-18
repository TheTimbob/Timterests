package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"timterests/internal/auth"
	"timterests/internal/storage"
)

// DownloadDocumentHandler handles document download requests for authenticated users.
func DownloadDocumentHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, key string, a *auth.Auth) {
	// Only admins can download documents
	if !a.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	if key == "" {
		http.Error(w, "Missing document key", http.StatusBadRequest)

		return
	}

	// Document listings use .yaml keys; serve the paired .md body file instead.
	if strings.HasSuffix(key, ".yaml") {
		key = strings.TrimSuffix(key, ".yaml") + ".md"
	}

	// Ensure the key is within the storage directory (prevents path traversal)
	localPath, err := storage.LocalPath(s.BaseDir, key)
	if err != nil {
		http.Error(w, "Invalid document key", http.StatusBadRequest)

		return
	}

	// Download from S3 if needed
	if s.UseS3 {
		err := s.DownloadS3File(r.Context(), key)
		if err != nil {
			log.Printf("DownloadDocumentHandler: failed to download from S3: %v", err)
			http.Error(w, "Failed to retrieve document", http.StatusInternalServerError)

			return
		}
	}

	fileName := filepath.Base(key)

	// Set headers to force download
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "text/markdown")

	http.ServeFile(w, r, localPath)
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
