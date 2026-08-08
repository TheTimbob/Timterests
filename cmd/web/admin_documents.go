package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"

	"timterests/internal/auth"
	apperrors "timterests/internal/errors"
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
	Index        int
}

// DocTypes returns the content directories the admin dashboard lists. A function
// rather than a package variable so callers cannot mutate the shared slice.
func DocTypes() []string {
	return []string{"articles", "projects", "reading-list", "letters"}
}

// AdminDocumentsParams holds the data passed to the admin documents template.
type AdminDocumentsParams struct {
	Docs       []DocumentInfo
	Query      string
	DocType    string // empty means all types
	From       string // YYYY-MM-DD, inclusive
	To         string // YYYY-MM-DD, inclusive
	SortBy     string
	SortDir    string
	Page       int
	TotalPages int
	Total      int
	DocTypes   []string
}

// AdminDocumentsPageHandler handles the admin documents dashboard at /admin/documents.
func AdminDocumentsPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	query := r.URL.Query().Get("q")
	docType := r.URL.Query().Get("type")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	sortBy := r.URL.Query().Get("sort")
	sortDir := r.URL.Query().Get("dir")
	pageStr := r.URL.Query().Get("page")

	// An unrecognised type would otherwise filter everything away with no clue why.
	if docType != "" && !slices.Contains(DocTypes(), docType) {
		docType = ""
	}

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
		HandleError(w, r, apperrors.StorageFailed(err), "AdminDocumentsPageHandler", "listDocuments")

		return
	}

	docs = filterDocs(docs, query, docType, from, to)
	sortDocs(docs, sortBy, sortDir)

	// Paginate
	total := len(docs)
	totalPages := max((total+docsPerPage-1)/docsPerPage, 1)

	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * docsPerPage
	end := min(start+docsPerPage, total)

	params := AdminDocumentsParams{
		Docs:       docs[start:end],
		Query:      query,
		DocType:    docType,
		From:       from,
		To:         to,
		SortBy:     sortBy,
		SortDir:    sortDir,
		Page:       page,
		TotalPages: totalPages,
		Total:      total,
		DocTypes:   DocTypes(),
	}

	var component templ.Component

	if IsHTMXRequest(r) {
		SetPartialResponseHeaders(w)

		component = AdminDocumentsTable(params)
	} else {
		component = AdminDocumentsPage(params)
	}

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "AdminDocumentsPageHandler", "render")
	}
}

// filterDocs narrows documents by filename search, document type, and modified
// date range. Empty values are ignored, so no filters returns everything.
func filterDocs(docs []DocumentInfo, query, docType, from, to string) []DocumentInfo {
	lower := strings.ToLower(query)
	// Dates that fail to parse are treated as absent rather than matching nothing.
	after, hasAfter := parseFilterDate(from)
	before, hasBefore := parseFilterDate(to)

	// "to" is inclusive, so compare against the end of that day.
	if hasBefore {
		before = before.AddDate(0, 0, 1)
	}

	filtered := docs[:0]

	for _, d := range docs {
		if lower != "" && !strings.Contains(strings.ToLower(d.Filename), lower) {
			continue
		}

		if docType != "" && d.DocType != docType {
			continue
		}

		if hasAfter && d.LastModified.Before(after) {
			continue
		}

		if hasBefore && !d.LastModified.Before(before) {
			continue
		}

		filtered = append(filtered, d)
	}

	return filtered
}

func parseFilterDate(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, false
	}

	return parsed, true
}

// sortDocs sorts documents in-place by the given field and direction.
func sortDocs(docs []DocumentInfo, sortBy, sortDir string) {
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
}

// ListAllDocuments collects DocumentInfo from all content type directories.
func ListAllDocuments(ctx context.Context, s storage.Storage) ([]DocumentInfo, error) {
	source := "Local"
	if s.UseS3 {
		source = "S3"
	}

	var docs []DocumentInfo

	for _, docType := range DocTypes() {
		objects, err := s.ListObjects(ctx, docType+"/")
		if err != nil {
			return nil, fmt.Errorf("listing %s: %w", docType, err)
		}

		for i, obj := range objects {
			key := aws.ToString(obj.Key)

			if key == "" || strings.HasSuffix(key, "/") {
				continue
			}

			docs = append(docs, DocumentInfo{
				Filename:     filepath.Base(key),
				Key:          key,
				DocType:      docType,
				Source:       source,
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
				Index:        i,
			})
		}
	}

	return docs, nil
}

// buildDocumentsURL constructs a properly encoded URL for the admin documents
// page, carrying the active filters through sorting and pagination. It takes the
// params struct so adding a filter does not mean touching every call site.
func buildDocumentsURL(p AdminDocumentsParams, sortBy, sortDir string, page int) string {
	v := url.Values{}

	for key, value := range map[string]string{
		"q":    p.Query,
		"type": p.DocType,
		"from": p.From,
		"to":   p.To,
	} {
		if value != "" {
			v.Set(key, value)
		}
	}

	v.Set("sort", sortBy)
	v.Set("dir", sortDir)
	v.Set("page", strconv.Itoa(page))

	return "/admin/documents?" + v.Encode()
}
