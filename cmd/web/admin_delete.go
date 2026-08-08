package web

import (
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"

	"timterests/internal/auth"
	apperrors "timterests/internal/errors"
	"timterests/internal/storage"
)

// errInvalidDocumentKey is returned for a key that does not name a document in a
// known content directory.
var errInvalidDocumentKey = errors.New("invalid document key")

// DeleteDocumentHandler removes a document and returns the refreshed table.
//
// POST only: a link or GET would let a crawler, prefetch or stray click destroy
// content.
func DeleteDocumentHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	if r.Method != http.MethodPost {
		HandleError(w, r, apperrors.MethodNotAllowed(), "DeleteDocumentHandler", "checkMethod")

		return
	}

	err := r.ParseForm()
	if err != nil {
		HandleError(w, r, apperrors.ParseFormFailed(err), "DeleteDocumentHandler", "parseForm")

		return
	}

	key := r.FormValue("key")
	if !validDocumentKey(key) {
		HandleError(w, r, apperrors.BadRequest(errInvalidDocumentKey), "DeleteDocumentHandler", "validateKey")

		return
	}

	err = s.DeleteDocument(r.Context(), key)
	if err != nil {
		log.Printf("delete: failed to delete %q: %v", key, err)
		HandleError(w, r, apperrors.StorageFailed(err), "DeleteDocumentHandler", "delete")

		return
	}

	// Re-render the table so the row disappears without a full page reload.
	AdminDocumentsPageHandler(w, r, s, a)
}

// validDocumentKey checks the key names a document in a known content directory.
// Without this, any path the storage layer accepts could be deleted through this
// endpoint.
func validDocumentKey(key string) bool {
	if !strings.HasSuffix(key, ".yaml") {
		return false
	}

	docType, rest, found := strings.Cut(key, "/")
	if !found || rest == "" || strings.Contains(rest, "/") {
		return false
	}

	return slices.Contains(DocTypes(), docType)
}
