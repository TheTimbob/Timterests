package web_test

import (
	"context"
	"testing"
	"timterests/cmd/web"
	"timterests/internal/service"
)

// TestArticleAdapters verifies that service-layer model types are correctly
// converted to Card components by the adapter functions.
func TestArticleAdapters(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("ArticleCard preserves all fields", func(t *testing.T) {
		ma, err := service.GetArticle(ctx, *s, "articles/test-article.yaml", 1)
		if err != nil {
			t.Fatalf("service.GetArticle failed: %v", err)
		}

		card := web.ArticleCard(*ma, 0)

		if card.Title != ma.Title {
			t.Errorf("Title mismatch: got %q, want %q", card.Title, ma.Title)
		}

		if card.Subtitle != ma.Subtitle {
			t.Errorf("Subtitle mismatch: got %q, want %q", card.Subtitle, ma.Subtitle)
		}

		if card.Date != ma.Date {
			t.Errorf("Date mismatch: got %q, want %q", card.Date, ma.Date)
		}
	})

	t.Run("ArticleCard from slice preserves length", func(t *testing.T) {
		mas, err := service.ListArticles(ctx, *s, "all")
		if err != nil {
			t.Fatalf("service.ListArticles failed: %v", err)
		}

		cards := make([]any, len(mas))
		for i, a := range mas {
			cards[i] = web.ArticleCard(a, i)
		}

		if len(cards) != len(mas) {
			t.Errorf("length mismatch: got %d, want %d", len(cards), len(mas))
		}
	})

	t.Run("ArticleCard produces correct URL", func(t *testing.T) {
		ma, err := service.GetArticle(ctx, *s, "articles/test-article.yaml", 5)
		if err != nil {
			t.Fatalf("service.GetArticle failed: %v", err)
		}

		card := web.ArticleCard(*ma, 0)

		expectedURL := "/article?id=5"
		if card.Get != expectedURL {
			t.Errorf("card URL: got %q, want %q", card.Get, expectedURL)
		}

		if card.ImagePath != "" {
			t.Errorf("article card should have no ImagePath, got %q", card.ImagePath)
		}

		if card.Date != ma.Date {
			t.Errorf("card date: got %q, want %q", card.Date, ma.Date)
		}
	})
}

// TestProjectAdapters verifies project card conversion.
func TestProjectAdapters(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("ProjectCard preserves all fields", func(t *testing.T) {
		mp, err := service.GetProject(ctx, *s, "projects/test-project.yaml", 2)
		if err != nil {
			t.Fatalf("service.GetProject failed: %v", err)
		}

		card := web.ProjectCard(*mp, 0)

		if card.Title != mp.Title {
			t.Errorf("Title mismatch: got %q, want %q", card.Title, mp.Title)
		}

		if card.ImagePath != mp.Image {
			t.Errorf("Image mismatch: got %q, want %q", card.ImagePath, mp.Image)
		}
	})

	t.Run("ProjectCard from slice preserves length", func(t *testing.T) {
		mps, err := service.ListProjects(ctx, *s, "all")
		if err != nil {
			t.Fatalf("service.ListProjects failed: %v", err)
		}

		cards := make([]any, len(mps))
		for i, p := range mps {
			cards[i] = web.ProjectCard(p, i)
		}

		if len(cards) != len(mps) {
			t.Errorf("length mismatch: got %d, want %d", len(cards), len(mps))
		}
	})

	t.Run("ProjectCard produces correct URL and ImagePath", func(t *testing.T) {
		mp, err := service.GetProject(ctx, *s, "projects/test-project.yaml", 3)
		if err != nil {
			t.Fatalf("service.GetProject failed: %v", err)
		}

		card := web.ProjectCard(*mp, 0)

		expectedURL := "/project?id=3"
		if card.Get != expectedURL {
			t.Errorf("card URL: got %q, want %q", card.Get, expectedURL)
		}

		if card.ImagePath != mp.Image {
			t.Errorf("card ImagePath: got %q, want %q", card.ImagePath, mp.Image)
		}

		expectedDate := mp.Timespan()
		if card.Date != expectedDate {
			t.Errorf("card Date: got %q, want %q", card.Date, expectedDate)
		}
	})
}

// TestLetterAdapters verifies letter card conversion.
func TestLetterAdapters(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("LetterCard preserves all fields", func(t *testing.T) {
		ml, err := service.GetLetter(ctx, *s, "letters/test-letter.yaml", 1)
		if err != nil {
			t.Fatalf("service.GetLetter failed: %v", err)
		}

		card := web.LetterCard(*ml, 0)

		if card.Title != ml.Title {
			t.Errorf("Title mismatch: got %q, want %q", card.Title, ml.Title)
		}

		if card.Date != ml.Date {
			t.Errorf("Date mismatch: got %q, want %q", card.Date, ml.Date)
		}
	})

	t.Run("LetterCard from slice preserves length", func(t *testing.T) {
		mls, err := service.ListLetters(ctx, *s, "all")
		if err != nil {
			t.Fatalf("service.ListLetters failed: %v", err)
		}

		cards := make([]any, len(mls))
		for i, l := range mls {
			cards[i] = web.LetterCard(l, i)
		}

		if len(cards) != len(mls) {
			t.Errorf("length mismatch: got %d, want %d", len(cards), len(mls))
		}
	})

	t.Run("LetterCard produces correct URL", func(t *testing.T) {
		ml, err := service.GetLetter(ctx, *s, "letters/test-letter.yaml", 4)
		if err != nil {
			t.Fatalf("service.GetLetter failed: %v", err)
		}

		card := web.LetterCard(*ml, 0)

		expectedURL := "/letter?id=4"
		if card.Get != expectedURL {
			t.Errorf("card URL: got %q, want %q", card.Get, expectedURL)
		}

		if card.ImagePath != "" {
			t.Errorf("letter card should have no ImagePath, got %q", card.ImagePath)
		}
	})
}

// TestReadingListAdapters verifies reading list card conversion.
func TestReadingListAdapters(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("BookCard preserves all fields", func(t *testing.T) {
		mb, err := service.GetBook(ctx, *s, "reading-list/test-book.yaml", 1)
		if err != nil {
			t.Fatalf("service.GetBook failed: %v", err)
		}

		card := web.BookCard(*mb, 0)

		if card.Title != mb.Title {
			t.Errorf("Title mismatch: got %q, want %q", card.Title, mb.Title)
		}

		if card.ImagePath != mb.Image {
			t.Errorf("Image mismatch: got %q, want %q", card.ImagePath, mb.Image)
		}
	})

	t.Run("BookCard from slice preserves length", func(t *testing.T) {
		mbs, err := service.ListBooks(ctx, *s, "all")
		if err != nil {
			t.Fatalf("service.ListBooks failed: %v", err)
		}

		cards := make([]any, len(mbs))
		for i, b := range mbs {
			cards[i] = web.BookCard(b, i)
		}

		if len(cards) != len(mbs) {
			t.Errorf("length mismatch: got %d, want %d", len(cards), len(mbs))
		}
	})

	t.Run("BookCard produces correct URL and ImagePath", func(t *testing.T) {
		mb, err := service.GetBook(ctx, *s, "reading-list/test-book.yaml", 7)
		if err != nil {
			t.Fatalf("service.GetBook failed: %v", err)
		}

		card := web.BookCard(*mb, 0)

		expectedURL := "/book?id=7"
		if card.Get != expectedURL {
			t.Errorf("card URL: got %q, want %q", card.Get, expectedURL)
		}

		if card.ImagePath != mb.Image {
			t.Errorf("card ImagePath: got %q, want %q", card.ImagePath, mb.Image)
		}

		if card.Date != "" {
			t.Errorf("book card should have no Date, got %q", card.Date)
		}
	})
}
