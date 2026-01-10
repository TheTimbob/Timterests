package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"timterests/cmd/web"
	"timterests/internal/auth"

	"github.com/PuerkitoBio/goquery"
)

func TestReadingListRendering(t *testing.T) {
	s := testSetup(t, context.Background())

	t.Run("render reading list page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/reading-list", nil)
		rec := httptest.NewRecorder()

		web.ReadingListPageHandler(rec, req, *s, "all", "list")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the title of the webpage to be present.
		if doc.Find("title").Length() == 0 {
			t.Error("expected title element to be rendered, but it wasn't")
		}
		// Expect the page name to be set correctly.
		categoryTitle := "Reading List"
		if actualPageName := doc.Find("h1.category-title").Text(); actualPageName != categoryTitle {
			t.Errorf("expected page name %q, got %q", categoryTitle, actualPageName)
		}
		// Expect the container element to be present.
		if doc.Find(`[id="reading-list-container"]`).Length() == 0 {
			t.Error("expected container element to be rendered, but it wasn't")
		}
	})

	t.Run("render reading list only", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/reading-list", nil)
		rec := httptest.NewRecorder()

		// Set the HX-Request header to trigger partial rendering
		req.Header.Set("Hx-Request", "true")

		web.ReadingListPageHandler(rec, req, *s, "all", "list")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the title of the webpage to not be rendered for the list.
		if doc.Find("title").Length() > 0 {
			t.Error("expected title element to not be rendered, but it was")
		}
		// Expect the page name element to not be rendered.
		if doc.Find("h1.category-title").Length() > 0 {
			t.Error("expected page name element to not be rendered, but it was")
		}
		// Expect the page-list to be rendered.
		if doc.Find(`ul#page-list`).Length() == 0 {
			t.Error("expected page-list element to be rendered, but it wasn't")
		}
	})

	t.Run("render books with a selected tag", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/reading-list", nil)
		rec := httptest.NewRecorder()

		tag := "Data Structures"
		web.ReadingListPageHandler(rec, req, *s, tag, "list")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the page-list to be rendered.
		if doc.Find(`ul#page-list`).Length() == 0 {
			t.Error("expected page-list element to be rendered, but it wasn't")
		}
		// Expect at least one book to be rendered.
		if doc.Find(`div.card-container`).Length() == 0 {
			t.Error("expected at least one book to be rendered, but none were")
		}
		// Expect the tag to be in the card tags.
		foundTag := false

		doc.Find(`p.card-tag`).Each(func(_ int, s *goquery.Selection) {
			if s.Text() == tag {
				foundTag = true
			}
		})

		if !foundTag {
			t.Errorf("expected to find tag %q in rendered cards", tag)
		}
	})

	t.Run("exclude books that don't have the selected tag", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/reading-list", nil)
		rec := httptest.NewRecorder()

		// Enter a non-existent tag to get zero results back (filter all books).
		tag := "non-existent-tag"
		web.ReadingListPageHandler(rec, req, *s, tag, "list")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the page-list to be rendered.
		if doc.Find(`ul#page-list`).Length() == 0 {
			t.Error("expected page-list element to be rendered, but it wasn't")
		}
		// Expect no books to be rendered.
		if doc.Find(`div.card-container`).Length() > 0 {
			t.Error("expected no books to be rendered, but some were")
		}
	})
}

func TestBookRendering(t *testing.T) {
	s := testSetup(t, context.Background())

	// Create auth instance for tests (won't be authenticated but prevents nil pointer)
	a := auth.NewAuth("test-session-key-minimum-32-bytes")

	t.Run("render book page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/book?id=0", nil)
		rec := httptest.NewRecorder()

		web.GetReadingListBook(rec, req, *s, "0", a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		// Expect the title of the webpage to be present
		if doc.Find("title").Length() == 0 {
			t.Error("expected title element to be rendered, but it wasn't")
		}

		// Expect the book content to be present
		if doc.Find("#reading-list-container").Length() == 0 {
			t.Error("expected book container to be rendered, but it wasn't")
		}

		// Expect headers to be present
		if doc.Find("h1").Length() == 0 {
			t.Error("expected h1 element to be rendered, but it wasn't")
		}

		if doc.Find("h2").Length() == 0 {
			t.Error("expected h2 element to be rendered, but it wasn't")
		}
	})

	t.Run("render book display only", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/book?id=0", nil)
		rec := httptest.NewRecorder()

		// Set HTMX header for partial rendering
		req.Header.Set("Hx-Request", "true")

		web.GetReadingListBook(rec, req, *s, "0", a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		// Expect the full page title to NOT be rendered
		if doc.Find("title").Length() > 0 {
			t.Error("expected title element to not be rendered, but it was")
		}

		// Expect the book content to be present
		if doc.Find("#reading-list-container").Length() == 0 {
			t.Error("expected book container to be rendered, but it wasn't")
		}

		// Expect headers to be present
		if doc.Find("h1").Length() == 0 {
			t.Error("expected h1 element to be rendered, but it wasn't")
		}

		if doc.Find("h2").Length() == 0 {
			t.Error("expected h2 element to be rendered, but it wasn't")
		}
	})
}

func TestReadingList(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("get book object", func(t *testing.T) {
		testBookPath := "reading-list/test-book.yaml"

		book, err := web.GetBook(ctx, testBookPath, 1, *s)
		if err != nil {
			t.Fatalf("failed to get book: %v", err)
		}

		if book.ID != "1" {
			t.Errorf("expected book ID '1', got '%s'", book.ID)
		}

		if book.Title != "Test Book" {
			t.Errorf("expected book title 'Test Book', got '%s'", book.Title)
		}

		if book.Author != "Test Author" {
			t.Errorf("expected book author 'Test Author', got '%s'", book.Author)
		}

		if book.Published != "2024" {
			t.Errorf("expected book published year '2024', got '%s'", book.Published)
		}

		if book.ISBN != "978-0-134685991" {
			t.Errorf("expected book ISBN '978-0-134685991', got '%s'", book.ISBN)
		}

		if book.Website != "https://example.com/test-book" {
			t.Errorf("expected book website 'https://example.com/test-book', got '%s'", book.Website)
		}

		if book.Status != "read" {
			t.Errorf("expected book status 'read', got '%s'", book.Status)
		}

		if book.Image == "" {
			t.Error("expected book image path to be set, but it was empty")
		}
	})

	t.Run("list books", func(t *testing.T) {
		books, err := web.ListBooks(ctx, *s, "")
		if err != nil {
			t.Fatalf("failed to list books: %v", err)
		}

		if len(books) < 1 {
			t.Errorf("expected at least one book, got %d", len(books))
		}
	})

	t.Run("list books with tag filter", func(t *testing.T) {
		books, err := web.ListBooks(ctx, *s, "Data Structures")
		if err != nil {
			t.Fatalf("failed to list books: %v", err)
		}

		if len(books) < 1 {
			t.Errorf("expected at least one book with tag 'Data Structures', got %d", len(books))
		}

		// Verify all returned books have the tag
		for _, book := range books {
			hasTag := slices.Contains(book.Tags, "Data Structures")

			if !hasTag {
				t.Errorf("book %q does not have tag 'Data Structures'", book.Title)
			}
		}
	})

	t.Run("book to card conversion", func(t *testing.T) {
		testBookPath := "reading-list/test-book.yaml"

		book, err := web.GetBook(ctx, testBookPath, 1, *s)
		if err != nil {
			t.Fatalf("failed to get book: %v", err)
		}

		card := book.ToCard(0)

		if card.Title != book.Title {
			t.Errorf("expected card title %q, got %q", book.Title, card.Title)
		}

		if card.Subtitle != book.Subtitle {
			t.Errorf("expected card subtitle %q, got %q", book.Subtitle, card.Subtitle)
		}

		if card.Get != "/book?id=1" {
			t.Errorf("expected card get URL '/book?id=1', got %q", card.Get)
		}

		// Books should have ImagePath but no Date
		if card.ImagePath == "" {
			t.Error("expected card image path to be set, but it was empty")
		}

		if card.Date != "" {
			t.Errorf("expected card date to be empty for books, got %q", card.Date)
		}

		if card.ImagePath != book.Image {
			t.Errorf("expected card image path %q, got %q", book.Image, card.ImagePath)
		}
	})
}
