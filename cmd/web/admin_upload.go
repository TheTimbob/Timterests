package web

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path"
	"slices"
	"strings"

	"timterests/internal/auth"
	apperrors "timterests/internal/errors"
	"timterests/internal/model"
	"timterests/internal/storage"

	"gopkg.in/yaml.v2"
)

// maxUploadBytes caps a single upload. Documents are text; anything larger is a
// mistake, and the request body is already limited to 10MB upstream.
const maxUploadBytes = 2 << 20

// UploadPageHandler renders the upload form.
func UploadPageHandler(w http.ResponseWriter, r *http.Request, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	err := renderHTML(w, r, http.StatusOK, UploadPage(UploadResult{DocTypes: DocTypes()}))
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "UploadPageHandler", "render")
	}
}

// UploadResult carries the outcome back to the template.
type UploadResult struct {
	DocTypes []string
	DocType  string
	Message  string
	Errors   []string
}

// UploadDocumentHandler accepts a .yaml/.md pair, validates the metadata against
// the chosen type, and writes both files.
//
// Validation runs before anything is written, so a document that would fail
// never lands half-uploaded.
func UploadDocumentHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	if r.Method != http.MethodPost {
		HandleError(w, r, apperrors.MethodNotAllowed(), "UploadDocumentHandler", "checkMethod")

		return
	}

	err := r.ParseMultipartForm(maxUploadBytes)
	if err != nil {
		HandleError(w, r, apperrors.ParseFormFailed(err), "UploadDocumentHandler", "parseForm")

		return
	}

	docType := r.FormValue("document-type")
	if !slices.Contains(DocTypes(), docType) {
		renderUpload(w, r, UploadResult{
			DocTypes: DocTypes(),
			Errors:   []string{"Choose a document type."},
		})

		return
	}

	result := uploadDocument(r, s, docType)

	renderUpload(w, r, result)
}

// uploadDocument does the work and reports what happened, so the handler stays
// about HTTP and this stays testable.
func uploadDocument(r *http.Request, s storage.Storage, docType string) UploadResult {
	result := UploadResult{DocTypes: DocTypes(), DocType: docType}

	yamlBytes, yamlName, err := readUpload(r, "yaml-file", ".yaml")
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	mdBytes, mdName, err := readUpload(r, "md-file", ".md")
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	if len(result.Errors) > 0 {
		return result
	}

	slug := strings.TrimSuffix(yamlName, ".yaml")
	if strings.TrimSuffix(mdName, ".md") != slug {
		result.Errors = append(result.Errors, fmt.Sprintf(
			"The two files must share a name: got %q and %q.", yamlName, mdName,
		))

		return result
	}

	missing := validateDocumentYAML(yamlBytes, docType)
	if missing != nil {
		result.Errors = append(result.Errors, missing.Error())

		return result
	}

	err = writeUploadedPair(r, s, docType, slug, yamlBytes, mdBytes)
	if err != nil {
		log.Printf("upload: failed to write %s/%s: %v", docType, slug, err)

		result.Errors = append(result.Errors, "Failed to save the document. Please try again.")

		return result
	}

	result.Message = fmt.Sprintf("Uploaded %s to %s.", slug, docType)

	return result
}

// readUpload pulls one file from the form and checks its extension.
func readUpload(r *http.Request, field, wantExt string) ([]byte, string, error) {
	file, header, err := r.FormFile(field)
	if err != nil {
		return nil, "", fmt.Errorf("a %s file is required", wantExt)
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			log.Printf("upload: failed to close %s: %v", field, closeErr)
		}
	}()

	// Strip any directory component, so a crafted filename cannot escape the
	// content directory.
	name := path.Base(header.Filename)
	if !strings.HasSuffix(name, wantExt) {
		return nil, "", fmt.Errorf("%q is not a %s file", name, wantExt)
	}

	content, err := readLimited(file)
	if err != nil {
		return nil, "", fmt.Errorf("%q %w", name, err)
	}

	return content, name, nil
}

// readLimited reads the whole file, refusing anything over the cap.
//
// A single Read is not enough: it may return fewer bytes than requested, which
// would silently truncate a document rather than fail. Reading one byte past the
// limit is how an oversized file is detected instead of quietly clipped.
func readLimited(file multipart.File) ([]byte, error) {
	content, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		return nil, errors.New("could not be read")
	}

	if len(content) > maxUploadBytes {
		return nil, fmt.Errorf("is larger than %dMB", maxUploadBytes>>20)
	}

	return content, nil
}

// validateDocumentYAML decodes the metadata into the struct for its type and
// checks the required fields. Decoding into the real type means the rules come
// from the struct tags, so a new document type is covered without changes here.
func validateDocumentYAML(content []byte, docType string) error {
	doc := emptyDocument(docType)
	if doc == nil {
		return errors.New("unsupported document type")
	}

	err := yaml.Unmarshal(content, doc)
	if err != nil {
		return fmt.Errorf("the YAML could not be parsed: %w", err)
	}

	err = model.ValidateRequired(doc)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// emptyDocument returns a pointer to the struct backing a document type.
func emptyDocument(docType string) any {
	switch docType {
	case "articles":
		return &model.Article{}
	case "projects":
		return &model.Project{}
	case "reading-list":
		return &model.ReadingList{}
	case "letters":
		return &model.Letter{}
	default:
		return nil
	}
}

func writeUploadedPair(
	r *http.Request,
	s storage.Storage,
	docType, slug string,
	yamlBytes, mdBytes []byte,
) error {
	for _, file := range []struct {
		key     string
		content []byte
	}{
		{docType + "/" + slug + ".yaml", yamlBytes},
		{docType + "/" + slug + ".md", mdBytes},
	} {
		err := s.WriteFile(r.Context(), file.key, file.content)
		if err != nil {
			return fmt.Errorf("writing %s: %w", file.key, err)
		}
	}

	return nil
}

func renderUpload(w http.ResponseWriter, r *http.Request, result UploadResult) {
	component := UploadPage(result)

	if IsHTMXRequest(r) {
		SetPartialResponseHeaders(w)

		component = UploadForm(result)
	}

	err := renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "renderUpload", "render")
	}
}
