package service_test

import (
	"context"
	"slices"
	"testing"
	"timterests/internal/service"
)

func TestListArticles(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("returns all articles when tag is empty", func(t *testing.T) {
		articles, err := service.ListArticles(ctx, *s, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(articles) < 1 {
			t.Errorf("expected at least one article, got %d", len(articles))
		}
	})

	t.Run("returns all articles when tag is 'all'", func(t *testing.T) {
		articles, err := service.ListArticles(ctx, *s, "all")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(articles) < 1 {
			t.Errorf("expected at least one article with tag 'all', got %d", len(articles))
		}
	})

	t.Run("filters articles by tag", func(t *testing.T) {
		articles, err := service.ListArticles(ctx, *s, "tag1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(articles) < 1 {
			t.Errorf("expected at least one article with tag 'tag1', got %d", len(articles))
		}

		for _, a := range articles {
			if !slices.Contains(a.Tags, "tag1") {
				t.Errorf("article %q does not have tag 'tag1'", a.Title)
			}
		}
	})

	t.Run("returns empty slice for non-existent tag", func(t *testing.T) {
		articles, err := service.ListArticles(ctx, *s, "does-not-exist")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(articles) != 0 {
			t.Errorf("expected zero articles for non-existent tag, got %d", len(articles))
		}
	})

	t.Run("articles are sorted by date descending", func(t *testing.T) {
		articles, err := service.ListArticles(ctx, *s, "all")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := 1; i < len(articles); i++ {
			if articles[i-1].Date < articles[i].Date {
				t.Errorf("articles not sorted by date desc: %q comes before %q", articles[i-1].Date, articles[i].Date)
			}
		}
	})
}

func TestGetArticle(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("retrieves article by key and id", func(t *testing.T) {
		article, err := service.GetArticle(ctx, *s, "articles/test-article.yaml", 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if article.ID != "1" {
			t.Errorf("expected ID '1', got %q", article.ID)
		}

		if article.Title != "Test Article" {
			t.Errorf("expected title 'Test Article', got %q", article.Title)
		}

		if article.Date == "" {
			t.Error("expected date to be populated, got empty string")
		}
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		_, err := service.GetArticle(ctx, *s, "articles/does-not-exist.yaml", 0)
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})
}

func TestGetLatestArticle(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("returns the most recent article", func(t *testing.T) {
		article, err := service.GetLatestArticle(ctx, *s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if article.Title == "" {
			t.Error("expected latest article title to be populated, got empty string")
		}

		if article.Date == "" {
			t.Error("expected latest article date to be populated, got empty string")
		}
	})
}

func TestFormatArticleDateForFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-01-15", "01-15-2026"},
		{"2024-12-31", "12-31-2024"},
		{"invalid", "invalid"}, // fallback: return as-is
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := service.FormatArticleDateForFilename(tc.input)
			if got != tc.expected {
				t.Errorf("FormatArticleDateForFilename(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
