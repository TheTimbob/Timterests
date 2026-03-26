package model_test

import (
	"testing"
	"timterests/internal/model"
)

func TestArticleValidate(t *testing.T) {
	t.Run("valid article passes validation", func(t *testing.T) {
		a := model.Article{
			Document: model.Document{Title: "My Article"},
			Date:     "2026-01-01",
		}
		if err := a.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("missing title fails validation", func(t *testing.T) {
		a := model.Article{Date: "2026-01-01"}
		if err := a.Validate(); err == nil {
			t.Error("expected error for missing title, got nil")
		}
	})

	t.Run("missing date fails validation", func(t *testing.T) {
		a := model.Article{Document: model.Document{Title: "My Article"}}
		if err := a.Validate(); err == nil {
			t.Error("expected error for missing date, got nil")
		}
	})
}

func TestProjectValidate(t *testing.T) {
	t.Run("valid project passes validation", func(t *testing.T) {
		p := model.Project{
			Document: model.Document{Title: "My Project"},
		}
		if err := p.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("missing title fails validation", func(t *testing.T) {
		p := model.Project{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for missing title, got nil")
		}
	})
}

func TestLetterValidate(t *testing.T) {
	t.Run("valid letter passes validation", func(t *testing.T) {
		l := model.Letter{
			Document: model.Document{Title: "Dear Friend"},
			Date:     "2026-01-01",
		}
		if err := l.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("missing title fails validation", func(t *testing.T) {
		l := model.Letter{Date: "2026-01-01"}
		if err := l.Validate(); err == nil {
			t.Error("expected error for missing title, got nil")
		}
	})

	t.Run("missing date fails validation", func(t *testing.T) {
		l := model.Letter{Document: model.Document{Title: "Dear Friend"}}
		if err := l.Validate(); err == nil {
			t.Error("expected error for missing date, got nil")
		}
	})
}

func TestReadingListValidate(t *testing.T) {
	t.Run("valid book passes validation", func(t *testing.T) {
		r := model.ReadingList{
			Document: model.Document{Title: "Go Programming"},
			Author:   "Alan Donovan",
		}
		if err := r.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("missing title fails validation", func(t *testing.T) {
		r := model.ReadingList{Author: "Alan Donovan"}
		if err := r.Validate(); err == nil {
			t.Error("expected error for missing title, got nil")
		}
	})

	t.Run("missing author fails validation", func(t *testing.T) {
		r := model.ReadingList{Document: model.Document{Title: "Go Programming"}}
		if err := r.Validate(); err == nil {
			t.Error("expected error for missing author, got nil")
		}
	})
}
