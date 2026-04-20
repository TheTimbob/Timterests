package service_test

import (
	"context"
	"slices"
	"testing"
	"timterests/internal/service"
)

func TestListBooks(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("returns all books when tag is empty", func(t *testing.T) {
		books, err := service.ListBooks(ctx, *s, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(books) < 1 {
			t.Errorf("expected at least one book, got %d", len(books))
		}
	})

	t.Run("returns all books when tag is 'all'", func(t *testing.T) {
		books, err := service.ListBooks(ctx, *s, "all")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(books) < 1 {
			t.Errorf("expected at least one book, got %d", len(books))
		}
	})

	t.Run("filters books by tag", func(t *testing.T) {
		books, err := service.ListBooks(ctx, *s, "Testing")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(books) < 1 {
			t.Errorf("expected at least one book with tag 'Testing', got %d", len(books))
		}

		for _, b := range books {
			if !slices.Contains(b.Tags, "Testing") {
				t.Errorf("book %q does not have tag 'Testing'", b.Title)
			}
		}
	})

	t.Run("returns empty slice for non-existent tag", func(t *testing.T) {
		books, err := service.ListBooks(ctx, *s, "does-not-exist")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(books) != 0 {
			t.Errorf("expected zero books for non-existent tag, got %d", len(books))
		}
	})
}

func TestGetBook(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("retrieves book by key and id", func(t *testing.T) {
		book, err := service.GetBook(ctx, *s, "reading-list/test-book.yaml", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if book.ID != "0" {
			t.Errorf("expected ID '0', got %q", book.ID)
		}

		if book.Title != "Test Book" {
			t.Errorf("expected title 'Test Book', got %q", book.Title)
		}

		if book.Author != "Test Author" {
			t.Errorf("expected author 'Test Author', got %q", book.Author)
		}
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		_, err := service.GetBook(ctx, *s, "reading-list/does-not-exist.yaml", 0)
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})

	t.Run("succeeds when book has no image path", func(t *testing.T) {
		book, err := service.GetBook(ctx, *s, "reading-list/no-image-book.yaml", 0)
		if err != nil {
			t.Fatalf("expected no error for book without image, got: %v", err)
		}

		if book.Image != "" {
			t.Errorf("expected empty image, got %q", book.Image)
		}

		if book.Title != "No Image Book" {
			t.Errorf("expected title 'No Image Book', got %q", book.Title)
		}
	})
}
