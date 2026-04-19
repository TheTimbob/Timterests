package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"timterests/internal/auth"
	apperrors "timterests/internal/errors"
	"timterests/internal/storage"
)

func DownloadDocumentHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, key string, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		HandleError(w, r, apperrors.Unauthorized(nil), "DownloadDocumentHandler", "auth")

		return
	}

	if key == "" {
		HandleError(w, r,
			apperrors.BadRequest(errors.New("missing document key")),
			"DownloadDocumentHandler", "validateKey")

		return
	}

	if base, ok := strings.CutSuffix(key, ".yaml"); ok {
		key = base + ".md"
	}

	localPath, err := storage.LocalPath(s.BaseDir, key)
	if err != nil {
		HandleError(w, r, apperrors.BadRequest(err), "DownloadDocumentHandler", "localPath")

		return
	}

	if s.UseS3 {
		err := s.DownloadS3File(r.Context(), key)
		if err != nil {
			HandleError(w, r, apperrors.StorageFailed(err), "DownloadDocumentHandler", "downloadS3")

			return
		}
	}

	fileName := filepath.Base(key)

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "text/markdown")

	http.ServeFile(w, r, localPath)
}

func DownloadNewDocumentHandler(w http.ResponseWriter, r *http.Request, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		HandleError(w, r, apperrors.Unauthorized(nil), "DownloadNewDocumentHandler", "auth")

		return
	}

	err := r.ParseForm()
	if err != nil {
		HandleError(w, r, apperrors.ParseFormFailed(err), "DownloadNewDocumentHandler", "parseForm")

		return
	}

	title := r.FormValue("title")
	subtitle := r.FormValue("subtitle")
	body := r.FormValue("body")

	if title == "" {
		HandleError(w, r,
			apperrors.BadRequest(errors.New("missing or invalid title field")),
			"DownloadNewDocumentHandler", "validateTitle")

		return
	}

	downloadFilename := storage.SanitizeFilename(title) + ".md"

	f, err := os.CreateTemp("", "download-*.md")
	if err != nil {
		HandleError(w, r, apperrors.InternalServerError(err), "DownloadNewDocumentHandler", "createTemp")

		return
	}

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
		HandleError(w, r, apperrors.InternalServerError(err), "DownloadNewDocumentHandler", "writeTemp")

		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+downloadFilename+"\"")
	w.Header().Set("Content-Type", "text/markdown")

	http.ServeFile(w, r, f.Name())
}
