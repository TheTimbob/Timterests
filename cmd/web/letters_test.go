package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"timterests/cmd/web"

	"github.com/PuerkitoBio/goquery"
)

func TestLetterListRendering(t *testing.T) {
	// Set up authentication once for all sub-tests
	a, addAuthCookie := testAuthentication(t)

	s := testSetup(t, context.Background())

	t.Run("render letter list page", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/letters", nil)
		rec := httptest.NewRecorder()

		// Add authentication cookie to this request
		addAuthCookie(req)

		web.LettersPageHandler(rec, req, *s, "all", "list", a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the title of the webpage to be present.
		if doc.Find("title").Length() == 0 {
			t.Error("expected title element to be rendered, but it wasn't")
		}
		// Expect the page name to be set correctly.
		categoryTitle := "Letters"
		if actualPageName := doc.Find("h1.category-title").Text(); actualPageName != categoryTitle {
			t.Errorf("expected page name %q, got %q", categoryTitle, actualPageName)
		}
		// Expect the container element to be present.
		if doc.Find(`[id="letters-container"]`).Length() == 0 {
			t.Error("expected container element to be rendered, but it wasn't")
		}
	})
	t.Run("render letter list only", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/letters", nil)
		rec := httptest.NewRecorder()

		// Add authentication cookie
		addAuthCookie(req)

		// Set the HX-Request header to trigger partial rendering
		req.Header.Set("Hx-Request", "true")

		web.LettersPageHandler(rec, req, *s, "all", "list", a)

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
	t.Run("render letters with a selected tag", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/letters", nil)
		rec := httptest.NewRecorder()

		// Add authentication cookie
		addAuthCookie(req)

		tag := "Tag1"
		web.LettersPageHandler(rec, req, *s, tag, "list", a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the page-list to be rendered.
		if doc.Find(`ul#page-list`).Length() == 0 {
			t.Error("expected page-list element to be rendered, but it wasn't")
		}
		// Expect at least one letter to be rendered.
		if doc.Find(`div.card-container`).Length() == 0 {
			t.Error("expected at least one letter to be rendered, but none were")
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

	t.Run("exclude letters that don't have the selected tag", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/letters", nil)
		rec := httptest.NewRecorder()

		// Add authentication cookie
		addAuthCookie(req)

		// Enter a non-existent tag to get zero results back (filter all letters).
		tag := "non-existent-tag"
		web.LettersPageHandler(rec, req, *s, tag, "list", a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the page-list to be rendered.
		if doc.Find(`ul#page-list`).Length() == 0 {
			t.Error("expected page-list element to be rendered, but it wasn't")
		}
		// Expect no letters to be rendered.
		if doc.Find(`div.card-container`).Length() > 0 {
			t.Error("expected no letters to be rendered, but some were")
		}
	})
}

func TestLetterRendering(t *testing.T) {
	s := testSetup(t, context.Background())

	// Set up authentication once for all sub-tests
	a, addAuthCookie := testAuthentication(t)

	t.Run("render letter page", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/letter?id=0", nil)
		rec := httptest.NewRecorder()

		// Add authentication cookie
		addAuthCookie(req)

		web.GetLetterHandler(rec, req, *s, "0", a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		// Expect the title of the webpage to be present
		if doc.Find("title").Length() == 0 {
			t.Error("expected title element to be rendered, but it wasn't")
		}

		// Expect the letter content to be present
		if doc.Find("#letter-container").Length() == 0 {
			t.Error("expected letter container to be rendered, but it wasn't")
		}

		// Expect headers to be present
		if doc.Find("h1").Length() == 0 {
			t.Error("expected h1 element to be rendered, but it wasn't")
		}

		if doc.Find("h2").Length() == 0 {
			t.Error("expected h2 element to be rendered, but it wasn't")
		}
	})

	t.Run("render letter display only", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/letter?id=0", nil)
		rec := httptest.NewRecorder()

		// Add authentication cookie
		addAuthCookie(req)

		// Set HTMX header for partial rendering
		req.Header.Set("Hx-Request", "true")

		web.GetLetterHandler(rec, req, *s, "0", a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		// Expect the full page title to NOT be rendered
		if doc.Find("title").Length() > 0 {
			t.Error("expected title element to not be rendered, but it was")
		}

		// Expect the letter content to be present
		if doc.Find("#letter-container").Length() == 0 {
			t.Error("expected letter container to be rendered, but it wasn't")
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

func TestLetter(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("get letter object", func(t *testing.T) {
		testLetterPath := "letters/test-letter.yaml"

		letter, err := web.GetLetter(ctx, testLetterPath, 1, *s)
		if err != nil {
			t.Fatalf("failed to get letter: %v", err)
		}

		if letter.ID != "1" {
			t.Errorf("expected letter ID '1', got '%s'", letter.ID)
		}

		if letter.Title != "Test Letter" {
			t.Errorf("expected letter title 'Test Letter', got '%s'", letter.Title)
		}

		if letter.Date != "2023-01-01" {
			t.Errorf("expected letter date '2023-01-01', got '%s'", letter.Date)
		}
	})

	t.Run("list letters", func(t *testing.T) {
		letters, err := web.ListLetters(ctx, *s, "")
		if err != nil {
			t.Fatalf("failed to list letters: %v", err)
		}

		if len(letters) < 1 {
			t.Errorf("expected at least one letter, got %d", len(letters))
		}
	})

	t.Run("list letters with tag filter", func(t *testing.T) {
		letters, err := web.ListLetters(ctx, *s, "Tag1")
		if err != nil {
			t.Fatalf("failed to list letters: %v", err)
		}

		if len(letters) < 1 {
			t.Errorf("expected at least one letter with tag 'Tag1', got %d", len(letters))
		}

		// Verify all returned letters have the tag
		for _, letter := range letters {
			hasTag := slices.Contains(letter.Tags, "Tag1")

			if !hasTag {
				t.Errorf("letter %q does not have tag 'Tag1'", letter.Title)
			}
		}
	})

	t.Run("letters are sorted by date descending", func(t *testing.T) {
		letters, err := web.ListLetters(ctx, *s, "")
		if err != nil {
			t.Fatalf("failed to list letters: %v", err)
		}

		if len(letters) < 2 {
			t.Skip("need at least 2 letters to test sorting")
		}

		// Verify letters are sorted by date in descending order (newest first)
		for i := range len(letters) - 1 {
			if letters[i].Date < letters[i+1].Date {
				t.Errorf("letters not sorted correctly: %s (%s) should come before %s (%s)",
					letters[i].Title, letters[i].Date, letters[i+1].Title, letters[i+1].Date)
			}
		}
	})

	t.Run("letter to card conversion", func(t *testing.T) {
		testLetterPath := "letters/test-letter.yaml"

		letter, err := web.GetLetter(ctx, testLetterPath, 1, *s)
		if err != nil {
			t.Fatalf("failed to get letter: %v", err)
		}

		card := letter.ToCard(0)

		if card.Title != letter.Title {
			t.Errorf("expected card title %q, got %q", letter.Title, card.Title)
		}

		if card.Subtitle != letter.Subtitle {
			t.Errorf("expected card subtitle %q, got %q", letter.Subtitle, card.Subtitle)
		}

		if card.Get != "/letter?id=1" {
			t.Errorf("expected card get URL '/letter?id=1', got %q", card.Get)
		}

		// Letters should have Date but no ImagePath
		if card.Date == "" {
			t.Error("expected card date to be set, but it was empty")
		}

		if card.ImagePath != "" {
			t.Errorf("expected card image path to be empty for letters, got %q", card.ImagePath)
		}

		if card.Date != letter.Date {
			t.Errorf("expected card date %q, got %q", letter.Date, card.Date)
		}
	})
}
