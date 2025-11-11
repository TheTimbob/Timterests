package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"timterests/internal/ai"
	"timterests/internal/storage"

	"github.com/a-h/templ"
)

const instructionFile = "prompts/article.txt"

func WriterPageHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, docType, key string, typeID int) {
	var content any
	var err error
	var component templ.Component

	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// If key is provided, get the content to load the existing document
	if key != "" {
		content, err = GetTypeContentFromID(docType, key, typeID, storageInstance)
		if err != nil {
			http.Error(w, "Failed to load document: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Create empty content based on docType
		switch docType {
		case "articles":
			content = &Article{}
		case "projects":
			content = &Project{}
		case "reading-list":
			content = &ReadingList{}
		case "letters":
			content = &Letter{}
		default:
			content = &Article{} // default to Article
		}
	}

	if r.Header.Get("HX-Request") == "true" && key == "" {
		component = FormContentByType(content)
	} else {
		component = WriterPage(content)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in WriterPage: %v", err)
	}
}

func WriteDocumentHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage) {

	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		log.Printf("Error parsing form: %v", err)
		return
	}

	// Check if S3 upload should be performed
	s3Upload := r.FormValue("s3-upload") == "on"
	delete(r.Form, "s3-upload")

	// Extract form data
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

	docType, ok := formData["document-type"].(string)
	if !ok || strings.TrimSpace(docType) == "" {
		http.Error(w, "Invalid or missing document type in form data", http.StatusBadRequest)
		log.Printf("Invalid or missing document type in form data")
		return
	}
	delete(formData, "document-type")

	// Create file name
	title, ok := formData["title"].(string)
	if !ok || strings.TrimSpace(title) == "" {
		http.Error(w, "Invalid or missing title in form data", http.StatusBadRequest)
		log.Printf("Invalid or missing title in form data")
		return
	}

	sanitizedTitle := storage.SanitizeFilename(title)

	var localFile string
	if docType == "articles" {
		articleDate := FormatDateForFilename(formData["date"].(string))
		localFile = fmt.Sprintf("%s-%s.yaml", sanitizedTitle, articleDate)
	} else {
		localFile = fmt.Sprintf("%s.yaml", sanitizedTitle)
	}

	// Create document
	localFilePath, err := storage.WriteYAMLDocument(localFile, formData)
	if err != nil {
		http.Error(w, "Failed to save document", http.StatusInternalServerError)
		log.Printf("Error writing document: %v", err)
		return
	}

	if s3Upload {
		// Upload to S3
		s3Path := docType + "/" + localFile
		err = storage.UploadFileToS3(r.Context(), storageInstance, s3Path, localFilePath)
		if err != nil {
			http.Error(w, "Failed to upload document to storage", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/writer", http.StatusSeeOther)
}

func WriterSuggestionHandler(w http.ResponseWriter, r *http.Request) {

	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	bodyContent := r.FormValue("body")
	if strings.TrimSpace(bodyContent) == "" {
		component := AISuggestionError("Please enter some content in the body field first.")
		err := component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Service temporarily unavailable", http.StatusBadRequest)
			log.Printf("Error rendering in WriterSuggestionHandler: %v", err)
		}
		return
	}

	suggestion, err := ai.GenerateSuggestion(r.Context(), bodyContent, instructionFile)
	if err != nil {
		component := AISuggestionError(fmt.Sprintf("Failed to get AI suggestion: %v", err))

		err = component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("Error rendering in WriterSuggestionHandler: %v", err)
		}
		return
	}

	component := AISuggestionResponse(suggestion)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func GetTypeContentFromID(docType, key string, id int, storageInstance storage.Storage) (any, error) {
	switch docType {
	case "articles":
		return GetArticle(key, id, storageInstance)
	case "projects":
		return GetProject(key, id, storageInstance)
	case "reading-list":
		return GetBook(key, id, storageInstance)
	case "letters":
		return GetLetter(key, id, storageInstance)
	default:
		return nil, fmt.Errorf("unsupported document type: %s", docType)
	}
}
