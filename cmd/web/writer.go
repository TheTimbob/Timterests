package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"timterests/internal/ai"
	"timterests/internal/auth"
	"timterests/internal/model"
	"timterests/internal/service"
	"timterests/internal/storage"

	"github.com/a-h/templ"
)

// WriterFormData holds all data needed to render the writer form.
type WriterFormData struct {
	Doc     model.Document
	Body    string
	DocType string
	Fields  templ.Component
}

// emptyFormData returns a zero-value WriterFormData for a new document of the given type.
func emptyFormData(docType string) WriterFormData {
	switch docType {
	case "projects":
		doc := model.Project{}

		return WriterFormData{Doc: doc.Document, DocType: "projects", Fields: ProjectFormContent(&doc)}
	case "reading-list":
		doc := model.ReadingList{}

		return WriterFormData{Doc: doc.Document, DocType: "reading-list", Fields: BookFormContent(&doc)}
	case "letters":
		doc := model.Letter{}

		return WriterFormData{Doc: doc.Document, DocType: "letters", Fields: LetterFormContent(&doc)}
	default: // "articles" and fallback
		doc := model.Article{}

		return WriterFormData{Doc: doc.Document, DocType: "articles", Fields: ArticleFormContent(&doc)}
	}
}

// WriterPageHandler handles requests to the writer page, ensuring authentication and rendering the appropriate content.
func WriterPageHandler(
	w http.ResponseWriter,
	r *http.Request,
	s storage.Storage,
	docType, key string,
	typeID int,
	a *auth.Auth) {
	var (
		data      WriterFormData
		err       error
		component templ.Component
	)

	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	if key != "" {
		data, err = getTypeContentRaw(r.Context(), docType, key, typeID, s)
		if err != nil {
			log.Printf("WriterPageHandler: failed to load document: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}
	} else {
		data = emptyFormData(docType)
	}

	if r.Header.Get("Hx-Request") == "true" && key == "" {
		component = WriterFormContent(data)
	} else {
		component = WriterPage(data)
	}

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		log.Printf("WriterPageHandler: failed to render: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WriteDocumentHandler handles the submission of the writer form to create or update documents.
func WriteDocumentHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	formData, s3Upload, err := extractFormData(r)
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		log.Printf("Error parsing form: %v", err)

		return
	}

	docType, err := extractDocType(formData)
	if err != nil {
		log.Printf("WriteDocumentHandler: invalid document type: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)

		return
	}

	slug, err := generateSlug(formData, docType)
	if err != nil {
		log.Printf("WriteDocumentHandler: invalid filename data: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)

		return
	}

	yamlFilename := docType + "/" + slug + ".yaml"
	mdFilename := docType + "/" + slug + ".md"

	yamlPath, err := storage.LocalPath(s.BaseDir, yamlFilename)
	if err != nil {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		log.Printf("Invalid local yaml path: %v", err)

		return
	}

	mdPath, err := storage.LocalPath(s.BaseDir, mdFilename)
	if err != nil {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		log.Printf("Invalid local md path: %v", err)

		return
	}

	err = storage.WriteMarkdownDocument(yamlPath, mdPath, formData)
	if err != nil {
		http.Error(w, "Failed to save document", http.StatusInternalServerError)
		log.Printf("Error writing document: %v", err)

		return
	}

	if s3Upload {
		err = s.UploadFileToS3(r.Context(), yamlFilename)
		if err != nil {
			http.Error(w, "Failed to upload document to storage", http.StatusInternalServerError)

			return
		}

		err = s.UploadFileToS3(r.Context(), mdFilename)
		if err != nil {
			http.Error(w, "Failed to upload document to storage", http.StatusInternalServerError)

			return
		}
	}

	http.Redirect(w, r, "/writer", http.StatusSeeOther)
}

// WriterSuggestionHandler handles AI-powered content suggestions for the writer.
func WriterSuggestionHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		log.Printf("Error parsing form in WriterSuggestionHandler: %v", err)

		return
	}

	bodyContent := r.FormValue("body")
	if strings.TrimSpace(bodyContent) == "" {
		component := AISuggestionError("Please enter some content in the body field first.")

		err := renderHTML(w, r, http.StatusOK, component)
		if err != nil {
			log.Printf("Error rendering in WriterSuggestionHandler: %v", err)
			http.Error(w, "Service temporarily unavailable", http.StatusInternalServerError)
		}

		return
	}

	docType := r.FormValue("document-type")
	if strings.TrimSpace(docType) == "" {
		docType = "articles" // default
	}

	systemInstruction, err := s.GetPromptContent(r.Context(), docType)
	if err != nil {
		log.Printf("Failed to load system prompt for docType %q: %v", docType, err) // #nosec G706

		component := AISuggestionError("AI suggestions are temporarily unavailable. Please try again later.")

		err = component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Service temporarily unavailable", http.StatusInternalServerError)
			log.Printf("Error rendering in WriterSuggestionHandler: %v", err)
		}

		return
	}

	suggestion, err := ai.GenerateSuggestion(r.Context(), bodyContent, systemInstruction)
	if err != nil {
		log.Printf("Failed to generate AI suggestion: %v", err)

		component := AISuggestionError("Failed to get AI suggestion. Please try again later.")

		err := renderHTML(w, r, http.StatusOK, component)
		if err != nil {
			log.Printf("Error rendering in WriterSuggestionHandler: %v", err)
			http.Error(w, "Service temporarily unavailable", http.StatusInternalServerError)
		}

		return
	}

	component := AISuggestionResponse(suggestion)

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		log.Printf("Error rendering AISuggestionResponse in WriterSuggestionHandler: %v", err)
		http.Error(w, "Service temporarily unavailable", http.StatusInternalServerError)

		return
	}
}

// metaSetter is satisfied by any type whose pointer embeds *model.Document.
type metaSetter interface {
	SetMeta(id, key string)
}

// loadRawDoc initialises a zero-value T, sets its metadata, fetches the raw
// (non-HTML-converted) file from storage, and returns Content with the doc and raw markdown body.
func loadRawDoc[T any, PT interface {
	*T
	metaSetter
}](ctx context.Context, key, idStr string, s storage.Storage) (model.Content[T], error) {
	var doc T
	PT(&doc).SetMeta(idStr, key)

	err := s.GetRawFile(ctx, key, PT(&doc))
	if err != nil {
		return model.Content[T]{}, fmt.Errorf("failed to get raw file: %w", err)
	}

	body, err := s.GetDocumentBodyRaw(ctx, key)
	if err != nil {
		log.Printf("loadRawDoc: failed to get raw body, leaving empty: %v", err)

		body = ""
	}

	return model.Content[T]{Doc: doc, Body: stripDocumentHeaders(body)}, nil
}

// stripDocumentHeaders removes the "# Title\n## Subtitle\n\n" prefix that
// WriteMarkdownDocument prepends, so the writer textarea shows only body content.
func stripDocumentHeaders(raw string) string {
	lines := strings.SplitN(raw, "\n", 4)
	if len(lines) >= 2 &&
		strings.HasPrefix(lines[0], "# ") &&
		strings.HasPrefix(lines[1], "## ") {
		if len(lines) == 4 {
			return lines[3]
		}

		return ""
	}

	return raw
}

// getTypeContentRaw retrieves content for editing, keeping the body as raw markdown.
func getTypeContentRaw(ctx context.Context, docType, key string, id int, s storage.Storage) (WriterFormData, error) {
	idStr := strconv.Itoa(id)

	switch docType {
	case "articles":
		c, err := loadRawDoc[model.Article](ctx, key, idStr, s)
		if err != nil {
			return WriterFormData{}, err
		}

		return WriterFormData{Doc: c.Doc.Document, Body: c.Body, DocType: "articles", Fields: ArticleFormContent(&c.Doc)}, nil
	case "projects":
		c, err := loadRawDoc[model.Project](ctx, key, idStr, s)
		if err != nil {
			return WriterFormData{}, err
		}

		return WriterFormData{Doc: c.Doc.Document, Body: c.Body, DocType: "projects", Fields: ProjectFormContent(&c.Doc)}, nil
	case "reading-list":
		c, err := loadRawDoc[model.ReadingList](ctx, key, idStr, s)
		if err != nil {
			return WriterFormData{}, err
		}

		return WriterFormData{Doc: c.Doc.Document, Body: c.Body, DocType: "reading-list",
			Fields: BookFormContent(&c.Doc)}, nil
	case "letters":
		c, err := loadRawDoc[model.Letter](ctx, key, idStr, s)
		if err != nil {
			return WriterFormData{}, err
		}

		return WriterFormData{Doc: c.Doc.Document, Body: c.Body, DocType: "letters", Fields: LetterFormContent(&c.Doc)}, nil
	default:
		return WriterFormData{}, fmt.Errorf("unsupported document type: %s", docType)
	}
}

// extractFormData parses form data and returns the processed data, S3 upload flag, and any error.
func extractFormData(r *http.Request) (map[string]any, bool, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse form: %w", err)
	}

	s3Upload := r.FormValue("s3-upload") == "on"
	delete(r.Form, "s3-upload")

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

	return formData, s3Upload, nil
}

// extractDocType validates and extracts the document type from form data.
func extractDocType(formData map[string]any) (string, error) {
	docTypeAny, ok := formData["document-type"]
	if !ok {
		return "", errors.New("missing document type in form data")
	}

	docType, ok := docTypeAny.(string)
	if !ok || strings.TrimSpace(docType) == "" {
		return "", errors.New("invalid or missing document type in form data")
	}

	delete(formData, "document-type")

	return docType, nil
}

// generateSlug creates a base filename slug (without extension) from the document type and form data.
func generateSlug(formData map[string]any, docType string) (string, error) {
	title, ok := formData["title"].(string)
	if !ok || strings.TrimSpace(title) == "" {
		return "", errors.New("invalid or missing title in form data")
	}

	sanitizedTitle := storage.SanitizeFilename(title)

	if docType == "articles" {
		articleDate, ok := formData["date"].(string)
		if !ok {
			return "", errors.New("invalid date in form data")
		}

		articleDate = service.FormatArticleDateForFilename(articleDate)

		return fmt.Sprintf("%s-%s", sanitizedTitle, articleDate), nil
	}

	return sanitizedTitle, nil
}
