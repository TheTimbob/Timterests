package model_test

import (
	"strings"
	"testing"

	"timterests/internal/model"
)

func TestValidateRequired(t *testing.T) {
	t.Parallel()

	t.Run("reports a field missing on the embedded Document", func(t *testing.T) {
		t.Parallel()

		err := model.ValidateRequired(&model.Article{Date: "2026-01-01"})
		if err == nil {
			t.Fatal("expected an error for a missing title")
		}

		if !strings.Contains(err.Error(), "title") {
			t.Errorf("expected the error to name title, got %q", err)
		}
	})

	t.Run("names every missing field, not just the first", func(t *testing.T) {
		t.Parallel()

		err := model.ValidateRequired(&model.Article{})
		if err == nil {
			t.Fatal("expected an error")
		}

		for _, want := range []string{"title", "date"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("expected %q in %q", want, err)
			}
		}
	})

	// Errors should use the yaml name, since that is what the uploaded file uses.
	t.Run("uses yaml names", func(t *testing.T) {
		t.Parallel()

		err := model.ValidateRequired(&model.ReadingList{})
		if err == nil {
			t.Fatal("expected an error")
		}

		if strings.Contains(err.Error(), "Author") {
			t.Errorf("expected the yaml name 'author', got %q", err)
		}
	})

	t.Run("passes when every required field is set", func(t *testing.T) {
		t.Parallel()

		complete := []any{
			&model.Article{Document: model.Document{Title: "t"}, Date: "2026-01-01"},
			&model.Letter{Document: model.Document{Title: "t"}, Date: "2026-01-01"},
			&model.Project{Document: model.Document{Title: "t"}},
			&model.ReadingList{Document: model.Document{Title: "t"}, Author: "a"},
		}

		for _, doc := range complete {
			err := model.ValidateRequired(doc)
			if err != nil {
				t.Errorf("expected %T to validate, got %v", doc, err)
			}
		}
	})

	// The point of the reflection approach: a type nobody wrote code for still
	// validates purely from its tags.
	t.Run("validates a type it has never seen", func(t *testing.T) {
		t.Parallel()

		type Recipe struct {
			model.Document `yaml:",inline"`

			Servings string `validate:"required" yaml:"servings"`
		}

		err := model.ValidateRequired(&Recipe{Document: model.Document{Title: "t"}})
		if err == nil || !strings.Contains(err.Error(), "servings") {
			t.Errorf("expected servings to be required, got %v", err)
		}
	})

	t.Run("tolerates nil and non-structs", func(t *testing.T) {
		t.Parallel()

		for _, value := range []any{nil, (*model.Article)(nil), "a string", 42} {
			err := model.ValidateRequired(value)
			if err != nil {
				t.Errorf("expected %v to be tolerated, got %v", value, err)
			}
		}
	})
}
