package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"timterests/cmd/web"

	"github.com/PuerkitoBio/goquery"
)

func TestAdminDocumentsPageHandler(t *testing.T) {
	a, addAuthCookie := testAuthentication(t)
	s := testSetup(t, context.Background())

	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/documents", nil)
		rec := httptest.NewRecorder()

		web.AdminDocumentsPageHandler(rec, req, *s, a)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
		}

		if loc := rec.Header().Get("Location"); loc != "/login" {
			t.Errorf("expected redirect to /login, got %q", loc)
		}
	})

	t.Run("renders full page when authenticated", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/documents", nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)

		web.AdminDocumentsPageHandler(rec, req, *s, a)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if doc.Find("title").Length() == 0 {
			t.Error("expected title element, but it wasn't found")
		}

		if doc.Find(`[id="admin-documents-container"]`).Length() == 0 {
			t.Error("expected admin-documents-container, but it wasn't found")
		}

		if doc.Find("h1.category-title").Text() != "Documents" {
			t.Errorf("expected page title 'Documents', got %q", doc.Find("h1.category-title").Text())
		}

		if doc.Find("table.admin-table").Length() == 0 {
			t.Error("expected admin-table, but it wasn't found")
		}
	})

	t.Run("renders partial table on HTMX request", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/documents", nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)
		req.Header.Set("Hx-Request", "true")

		web.AdminDocumentsPageHandler(rec, req, *s, a)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if doc.Find("title").Length() > 0 {
			t.Error("expected no title element for partial render, but found one")
		}

		if doc.Find(`[id="documents-table-wrapper"]`).Length() == 0 {
			t.Error("expected documents-table-wrapper, but it wasn't found")
		}

		if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-store") {
			t.Errorf("expected Cache-Control to contain no-store, got %q", cc)
		}

		if vary := rec.Header().Get("Vary"); !strings.Contains(vary, "HX-Request") {
			t.Errorf("expected Vary to contain HX-Request, got %q", vary)
		}
	})

	t.Run("lists documents from all content types", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/documents", nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)

		web.AdminDocumentsPageHandler(rec, req, *s, a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		// Expect at least one row in the table body
		if doc.Find("table.admin-table tbody tr").Length() == 0 {
			t.Error("expected at least one document row, but found none")
		}
	})

	t.Run("search filters by filename", func(t *testing.T) {
		// Get unfiltered row count to compare against
		unfilteredReq := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/documents", nil)
		unfilteredRec := httptest.NewRecorder()

		addAuthCookie(unfilteredReq)
		web.AdminDocumentsPageHandler(unfilteredRec, unfilteredReq, *s, a)

		unfilteredDoc, err := goquery.NewDocumentFromReader(unfilteredRec.Body)
		if err != nil {
			t.Fatalf("failed to parse unfiltered response: %v", err)
		}

		totalRows := 0

		unfilteredDoc.Find("table.admin-table tbody tr").Each(func(_ int, row *goquery.Selection) {
			if row.Find("td").Length() > 1 {
				totalRows++
			}
		})

		// Search for a query that matches a subset of documents
		const query = "test-article"

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/documents?q="+query, nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)
		web.AdminDocumentsPageHandler(rec, req, *s, a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		matchingRows := 0

		doc.Find("table.admin-table tbody tr").Each(func(_ int, row *goquery.Selection) {
			if row.Find("td").Length() > 1 {
				filename := row.Find("td").First().Text()
				if !strings.Contains(strings.ToLower(filename), query) {
					t.Errorf("filename %q does not contain query %q", filename, query)
				}

				matchingRows++
			}
		})

		if matchingRows == 0 {
			t.Error("expected at least one matching row, got none")
		}

		if matchingRows >= totalRows {
			t.Errorf("expected filtered rows (%d) to be fewer than total rows (%d)", matchingRows, totalRows)
		}
	})

	t.Run("sort parameter is respected", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodGet, "/admin/documents?sort=filename&dir=asc", nil,
		)
		rec := httptest.NewRecorder()

		addAuthCookie(req)
		web.AdminDocumentsPageHandler(rec, req, *s, a)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})
}

func TestAdminDocumentsSortOrder(t *testing.T) {
	a, addAuthCookie := testAuthentication(t)
	s := testSetup(t, context.Background())

	t.Run("filename ascending is ordered", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodGet, "/admin/documents?sort=filename&dir=asc", nil,
		)
		rec := httptest.NewRecorder()

		addAuthCookie(req)
		web.AdminDocumentsPageHandler(rec, req, *s, a)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		var filenames []string

		doc.Find("table.admin-table tbody tr td:first-child").Each(func(_ int, sel *goquery.Selection) {
			name := strings.TrimSpace(sel.Text())
			if name != "" {
				filenames = append(filenames, name)
			}
		})

		if len(filenames) < 2 {
			t.Fatalf("expected at least 2 filenames to verify sort, got %d", len(filenames))
		}

		for i := 1; i < len(filenames); i++ {
			if filenames[i-1] > filenames[i] {
				t.Errorf("filenames not sorted ascending: %q > %q", filenames[i-1], filenames[i])
			}
		}
	})

	t.Run("filename descending is ordered", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodGet, "/admin/documents?sort=filename&dir=desc", nil,
		)
		rec := httptest.NewRecorder()

		addAuthCookie(req)
		web.AdminDocumentsPageHandler(rec, req, *s, a)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		var filenames []string

		doc.Find("table.admin-table tbody tr td:first-child").Each(func(_ int, sel *goquery.Selection) {
			name := strings.TrimSpace(sel.Text())
			if name != "" {
				filenames = append(filenames, name)
			}
		})

		if len(filenames) < 2 {
			t.Fatalf("expected at least 2 filenames to verify sort, got %d", len(filenames))
		}

		for i := 1; i < len(filenames); i++ {
			if filenames[i-1] < filenames[i] {
				t.Errorf("filenames not sorted descending: %q < %q", filenames[i-1], filenames[i])
			}
		}
	})

	t.Run("modified date sort returns 200", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodGet, "/admin/documents?sort=modified&dir=desc", nil,
		)
		rec := httptest.NewRecorder()

		addAuthCookie(req)
		web.AdminDocumentsPageHandler(rec, req, *s, a)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})
}

func TestAdminDocumentsSortByModified(t *testing.T) {
	a, addAuthCookie := testAuthentication(t)
	s := testSetup(t, context.Background())

	t.Run("modified ascending is ordered", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodGet,
			"/admin/documents?sort=modified&dir=asc", nil,
		)
		rec := httptest.NewRecorder()

		addAuthCookie(req)
		web.AdminDocumentsPageHandler(rec, req, *s, a)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		var dates []string

		doc.Find("table.admin-table tbody tr").Each(func(_ int, row *goquery.Selection) {
			date := strings.TrimSpace(row.Find("td").Eq(4).Text())
			if date != "" {
				dates = append(dates, date)
			}
		})

		if len(dates) < 2 {
			t.Fatalf("expected at least 2 dates to verify sort, got %d", len(dates))
		}

		for i := 1; i < len(dates); i++ {
			if dates[i-1] > dates[i] {
				t.Errorf("dates not sorted ascending: %q > %q", dates[i-1], dates[i])
			}
		}
	})

	t.Run("modified descending is ordered", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodGet,
			"/admin/documents?sort=modified&dir=desc", nil,
		)
		rec := httptest.NewRecorder()

		addAuthCookie(req)
		web.AdminDocumentsPageHandler(rec, req, *s, a)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		var dates []string

		doc.Find("table.admin-table tbody tr").Each(func(_ int, row *goquery.Selection) {
			date := strings.TrimSpace(row.Find("td").Eq(4).Text())
			if date != "" {
				dates = append(dates, date)
			}
		})

		if len(dates) < 2 {
			t.Fatalf("expected at least 2 dates to verify sort, got %d", len(dates))
		}

		for i := 1; i < len(dates); i++ {
			if dates[i-1] < dates[i] {
				t.Errorf("dates not sorted descending: %q < %q", dates[i-1], dates[i])
			}
		}
	})
}

func TestAdminDocumentsFilters(t *testing.T) {
	a, addAuthCookie := testAuthentication(t)
	s := testSetup(t, context.Background())

	// rowsFor returns the Type column of every listed document.
	rowsFor := func(t *testing.T, query string) []string {
		t.Helper()

		req := httptest.NewRequestWithContext(
			context.Background(), http.MethodGet, "/admin/documents"+query, nil,
		)
		rec := httptest.NewRecorder()

		addAuthCookie(req)
		web.AdminDocumentsPageHandler(rec, req, *s, a)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		var types []string

		doc.Find("table.admin-table tbody tr").Each(func(_ int, row *goquery.Selection) {
			// Skip the "No documents found." placeholder row.
			if row.Find(".admin-table-empty").Length() > 0 {
				return
			}

			types = append(types, strings.TrimSpace(row.Find("td").Eq(1).Text()))
		})

		return types
	}

	unfiltered := rowsFor(t, "")
	if len(unfiltered) < 2 {
		t.Fatalf("need at least 2 documents to test filtering, got %d", len(unfiltered))
	}

	t.Run("type filter returns only that type", func(t *testing.T) {
		types := rowsFor(t, "?type=articles")

		if len(types) == 0 {
			t.Fatal("expected at least one article")
		}

		for _, got := range types {
			if got != "articles" {
				t.Errorf("expected only articles, got %q", got)
			}
		}

		if len(types) >= len(unfiltered) {
			t.Error("expected the type filter to narrow the list")
		}
	})

	// An unknown type is ignored rather than silently hiding everything.
	t.Run("unknown type is ignored", func(t *testing.T) {
		if got := len(rowsFor(t, "?type=nonsense")); got != len(unfiltered) {
			t.Errorf("expected all %d documents, got %d", len(unfiltered), got)
		}
	})

	t.Run("future date range excludes everything", func(t *testing.T) {
		if got := len(rowsFor(t, "?from=2999-01-01")); got != 0 {
			t.Errorf("expected no documents modified after 2999, got %d", got)
		}
	})

	// A malformed date must not be treated as a filter that matches nothing.
	t.Run("unparseable date is ignored", func(t *testing.T) {
		if got := len(rowsFor(t, "?from=not-a-date")); got != len(unfiltered) {
			t.Errorf("expected all %d documents, got %d", len(unfiltered), got)
		}
	})

	t.Run("filters survive sorting links", func(t *testing.T) {
		types := rowsFor(t, "?type=articles&sort=modified&dir=desc")
		for _, got := range types {
			if got != "articles" {
				t.Errorf("expected filter to persist through sorting, got %q", got)
			}
		}
	})
}

func TestListAllDocuments(t *testing.T) {
	s := testSetup(t, context.Background())

	t.Run("returns documents from all content types", func(t *testing.T) {
		docs, err := web.ListAllDocuments(context.Background(), *s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(docs) == 0 {
			t.Error("expected documents, got none")
		}

		// Check that multiple doc types are represented
		types := make(map[string]bool)
		for _, d := range docs {
			types[d.DocType] = true
		}

		if len(types) < 2 {
			t.Errorf("expected documents from multiple types, got: %v", types)
		}
	})
}
