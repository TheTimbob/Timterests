package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"testing"
	"timterests/cmd/web"
	"timterests/internal/storage"

	"github.com/PuerkitoBio/goquery"
)

func testSetup(t *testing.T, ctx context.Context) *storage.Storage {
	t.Helper()
	t.Setenv("USE_S3", "false")

	s, err := storage.NewStorage(ctx)
	if err != nil {
		t.Fatalf("failed to initialize storage: %v", err)
	}

	s.BaseDir = filepath.Join(s.BaseDir, "testdata")

	return s
}

func TestArticleListRendering(t *testing.T) {
	s := testSetup(t, context.Background())

	t.Run("render article list page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)
		rec := httptest.NewRecorder()

		web.ArticlesPageHandler(rec, req, *s, "all", "list")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the title of the webpage to be present.
		if doc.Find("title").Length() == 0 {
			t.Error("expected title element to be rendered, but it wasn't")
		}
		// Expect the page name to be set correctly.
		categoryTitle := "Articles"
		if actualPageName := doc.Find("h1.category-title").Text(); actualPageName != categoryTitle {
			t.Errorf("expected page name %q, got %q", categoryTitle, actualPageName)
		}
		// Expect the container element to be present.
		if doc.Find(`[id="articles-container"]`).Length() == 0 {
			t.Error("expected container element to be rendered, but it wasn't")
		}
	})
	t.Run("render article list only", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)
		rec := httptest.NewRecorder()

		// Set the HX-Request header to trigger partial rendering
		req.Header.Set("Hx-Request", "true")

		web.ArticlesPageHandler(rec, req, *s, "all", "list")

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
	t.Run("render articles with a selected tag", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)
		rec := httptest.NewRecorder()

		tag := "tag1"
		web.ArticlesPageHandler(rec, req, *s, tag, "list")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the page-list to be rendered.
		if doc.Find(`ul#page-list`).Length() == 0 {
			t.Error("expected page-list element to be rendered, but it wasn't")
		}
		// Expect at least one article to be rendered.
		if doc.Find(`div.card-container`).Length() == 0 {
			t.Error("expected at least one article to be rendered, but none were")
		}
		// Expect the tag to contain the body of the tag filter.
		if actualTag := doc.Find(`p.card-tag`).First().Text(); actualTag != tag {
			t.Errorf("expected selected tag %q, got %q", tag, actualTag)
		}
	})

	t.Run("exclude articles that have the selected tag", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/articles", nil)
		rec := httptest.NewRecorder()

		// Enter a non-existent tag to get zero results back (filter all articles).
		tag := "non-existent-tag"
		web.ArticlesPageHandler(rec, req, *s, tag, "list")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the page-list to be rendered.
		if doc.Find(`ul#page-list`).Length() == 0 {
			t.Error("expected page-list element to be rendered, but it wasn't")
		}
		// Expect no articles to be rendered.
		if doc.Find(`div.card-container`).Length() > 0 {
			t.Error("expected no articles to be rendered, but some were")
		}
	})
}

func TestArticleRendering(t *testing.T) {
	s := testSetup(t, context.Background())

	t.Run("render article page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/article?id=0", nil)
		rec := httptest.NewRecorder()

		web.GetArticleHandler(rec, req, *s, "0")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		// Expect the title of the webpage to be present
		if doc.Find("title").Length() == 0 {
			t.Error("expected title element to be rendered, but it wasn't")
		}

		// Expect the article content to be present
		if doc.Find("#article-container").Length() == 0 {
			t.Error("expected article container to be rendered, but it wasn't")
		}

		// Expect headers to be present
		if doc.Find("h1").Length() == 0 {
			t.Error("expected h1 element to be rendered, but it wasn't")
		}

		if doc.Find("h2").Length() == 0 {
			t.Error("expected h2 element to be rendered, but it wasn't")
		}
	})

	t.Run("render article display only", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/article?id=0", nil)
		rec := httptest.NewRecorder()

		// Set HTMX header for partial rendering
		req.Header.Set("Hx-Request", "true")

		web.GetArticleHandler(rec, req, *s, "0")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		// Expect the full page title to NOT be rendered
		if doc.Find("title").Length() > 0 {
			t.Error("expected title element to not be rendered, but it was")
		}

		// Expect the article content to be present
		if doc.Find("#article-container").Length() == 0 {
			t.Error("expected article container to be rendered, but it wasn't")
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

func TestArticle(t *testing.T) {
	context := context.Background()
	s := testSetup(t, context)

	t.Run("get article object", func(t *testing.T) {
		testArticle := "articles/test-article.yaml"

		article, err := web.GetArticle(context, testArticle, 1, *s)
		if err != nil {
			t.Fatalf("failed to get article: %v", err)
		}

		if article.ID != "1" {
			t.Errorf("expected article ID '1', got '%s'", article.ID)
		}

		if article.Title != "Test Article" {
			t.Errorf("expected article title 'Test Article', got '%s'", article.Title)
		}
	})

	t.Run("list articles", func(t *testing.T) {
		articles, err := web.ListArticles(context, *s, "")
		if err != nil {
			t.Fatalf("failed to list articles: %v", err)
		}

		if len(articles) < 1 {
			t.Errorf("expected at least one article, got %d", len(articles))
		}
	})

	t.Run("list articles with tag filter", func(t *testing.T) {
		articles, err := web.ListArticles(context, *s, "tag1")
		if err != nil {
			t.Fatalf("failed to list articles: %v", err)
		}

		if len(articles) < 1 {
			t.Errorf("expected at least one article with tag 'tag1', got %d", len(articles))
		}

		// Verify all returned articles have the tag
		for _, article := range articles {
			hasTag := slices.Contains(article.Tags, "tag1")

			if !hasTag {
				t.Errorf("article %q does not have tag 'tag1'", article.Title)
			}
		}
	})

	t.Run("get latest article", func(t *testing.T) {
		article, err := web.GetLatestArticle(context, *s)
		if err != nil {
			t.Fatalf("failed to get latest article: %v", err)
		}

		expectedTitle := "Test Article"
		if article.Title != expectedTitle {
			t.Errorf("expected latest article title %q, got %q", expectedTitle, article.Title)
		}
	})

	t.Run("article to card conversion", func(t *testing.T) {
		testArticle := "articles/test-article.yaml"

		article, err := web.GetArticle(context, testArticle, 1, *s)
		if err != nil {
			t.Fatalf("failed to get article: %v", err)
		}

		card := article.ToCard(0)

		if card.Title != article.Title {
			t.Errorf("expected card title %q, got %q", article.Title, card.Title)
		}

		if card.Subtitle != article.Subtitle {
			t.Errorf("expected card subtitle %q, got %q", article.Subtitle, card.Subtitle)
		}

		if card.Get != "/article?id=1" {
			t.Errorf("expected card get URL '/article?id=1', got %q", card.Get)
		}
	})
}
