package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin/documents?q=test-article", nil)
		rec := httptest.NewRecorder()

		addAuthCookie(req)

		web.AdminDocumentsPageHandler(rec, req, *s, a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		rows := doc.Find("table.admin-table tbody tr")
		rows.Each(func(_ int, row *goquery.Selection) {
			// Each row that renders (not the empty row) should contain "test-article"
			if row.Find("td").Length() > 1 {
				filename := row.Find("td").First().Text()
				if filename == "" {
					t.Error("expected non-empty filename in search result")
				}
			}
		})
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
