package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"

	"timterests/internal/auth"
	"timterests/internal/storage"
)

const docsPerPage = 20

// DocumentInfo holds metadata about a stored document.
type DocumentInfo struct {
	Filename     string
	Key          string
	DocType      string
	Source       string
	Size         int64
	LastModified time.Time
}

// AdminDocumentsParams holds the data passed to the admin documents template.
type AdminDocumentsParams struct {
	Docs       []DocumentInfo
	Query      string
	SortBy     string
	SortDir    string
	Page       int
	TotalPages int
	Total      int
}

// AdminDocumentsPageHandler handles the admin documents dashboard at /admin/documents.
func AdminDocumentsPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	query := r.URL.Query().Get("q")
	sortBy := r.URL.Query().Get("sort")
	sortDir := r.URL.Query().Get("dir")
	pageStr := r.URL.Query().Get("page")

	if sortBy == "" {
		sortBy = "filename"
	}

	if sortDir == "" {
		sortDir = "asc"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	docs, err := ListAllDocuments(r.Context(), s)
	if err != nil {
		http.Error(w, "Failed to list documents", http.StatusInternalServerError)
		log.Printf("Error listing documents: %v", err)

		return
	}

	// Search by filename (case-insensitive)
	if query != "" {
		lower := strings.ToLower(query)
		filtered := docs[:0]

		for _, d := range docs {
			if strings.Contains(strings.ToLower(d.Filename), lower) {
				filtered = append(filtered, d)
			}
		}

		docs = filtered
	}

	// Sort
	sort.Slice(docs, func(i, j int) bool {
		switch sortBy {
		case "modified":
			if sortDir == "desc" {
				return docs[i].LastModified.After(docs[j].LastModified)
			}

			return docs[i].LastModified.Before(docs[j].LastModified)
		default: // filename
			if sortDir == "desc" {
				return docs[i].Filename > docs[j].Filename
			}

			return docs[i].Filename < docs[j].Filename
		}
	})

	// Paginate
	total := len(docs)
	totalPages := (total + docsPerPage - 1) / docsPerPage

	if totalPages == 0 {
		totalPages = 1
	}

	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * docsPerPage
	end := start + docsPerPage

	if end > total {
		end = total
	}

	params := AdminDocumentsParams{
		Docs:       docs[start:end],
		Query:      query,
		SortBy:     sortBy,
		SortDir:    sortDir,
		Page:       page,
		TotalPages: totalPages,
		Total:      total,
	}

	var component templ.Component

	if r.Header.Get("Hx-Request") == "true" {
		component = AdminDocumentsTable(params)
	} else {
		component = AdminDocumentsPage(params)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering AdminDocumentsPage: %v", err)
	}
}

// ListAllDocuments collects DocumentInfo from all content type directories.
func ListAllDocuments(ctx context.Context, s storage.Storage) ([]DocumentInfo, error) {
	docTypes := []string{"articles", "projects", "reading-list", "letters"}

	source := "Local"
	if s.UseS3 {
		source = "S3"
	}

	var docs []DocumentInfo

	for _, docType := range docTypes {
		objects, err := s.ListObjects(ctx, docType+"/")
		if err != nil {
			return nil, fmt.Errorf("listing %s: %w", docType, err)
		}

		for _, obj := range objects {
			key := aws.ToString(obj.Key)

			docs = append(docs, DocumentInfo{
				Filename:     filepath.Base(key),
				Key:          key,
				DocType:      docType,
				Source:       source,
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
			})
		}
	}

	return docs, nil
}

// formatFileSize formats a byte count as a human-readable string.
func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}

	return fmt.Sprintf("%.1f KB", float64(size)/1024)
}

// buildDocumentsURL constructs a properly encoded URL for the admin documents page.
func buildDocumentsURL(query, sortBy, sortDir string, page int) string {
	v := url.Values{}

	if query != "" {
		v.Set("q", query)
	}

	v.Set("sort", sortBy)
	v.Set("dir", sortDir)
	v.Set("page", strconv.Itoa(page))

	return "/admin/documents?" + v.Encode()
}
