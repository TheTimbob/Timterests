package service_test

import (
	"context"
	"slices"
	"testing"
	"timterests/internal/service"
)

func TestListLetters(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("returns all letters when tag is empty", func(t *testing.T) {
		letters, err := service.ListLetters(ctx, *s, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(letters) < 1 {
			t.Errorf("expected at least one letter, got %d", len(letters))
		}
	})

	t.Run("returns all letters when tag is 'all'", func(t *testing.T) {
		letters, err := service.ListLetters(ctx, *s, "all")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(letters) < 1 {
			t.Errorf("expected at least one letter, got %d", len(letters))
		}
	})

	t.Run("filters letters by tag", func(t *testing.T) {
		letters, err := service.ListLetters(ctx, *s, "Tag1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(letters) < 1 {
			t.Errorf("expected at least one letter with tag 'Tag1', got %d", len(letters))
		}

		for _, l := range letters {
			if !slices.Contains(l.Tags, "Tag1") {
				t.Errorf("letter %q does not have tag 'Tag1'", l.Title)
			}
		}
	})

	t.Run("returns empty slice for non-existent tag", func(t *testing.T) {
		letters, err := service.ListLetters(ctx, *s, "does-not-exist")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(letters) != 0 {
			t.Errorf("expected zero letters for non-existent tag, got %d", len(letters))
		}
	})

	t.Run("letters are sorted by date descending", func(t *testing.T) {
		letters, err := service.ListLetters(ctx, *s, "all")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := 1; i < len(letters); i++ {
			if letters[i-1].Date < letters[i].Date {
				t.Errorf("letters not sorted by date desc: %q comes before %q", letters[i-1].Date, letters[i].Date)
			}
		}
	})
}

func TestGetLetter(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("retrieves letter by key and id", func(t *testing.T) {
		letter, err := service.GetLetter(ctx, *s, "letters/test-letter.md", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if letter.ID != "0" {
			t.Errorf("expected ID '0', got %q", letter.ID)
		}

		if letter.Title != "Test Letter" {
			t.Errorf("expected title 'Test Letter', got %q", letter.Title)
		}
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		_, err := service.GetLetter(ctx, *s, "letters/does-not-exist.md", 0)
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})
}
