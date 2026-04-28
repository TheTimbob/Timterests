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

		err := a.Validate()
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("missing title fails validation", func(t *testing.T) {
		a := model.Article{Date: "2026-01-01"}

		err := a.Validate()
		if err == nil {
			t.Error("expected error for missing title, got nil")
		}
	})

	t.Run("missing date fails validation", func(t *testing.T) {
		a := model.Article{Document: model.Document{Title: "My Article"}}

		err := a.Validate()
		if err == nil {
			t.Error("expected error for missing date, got nil")
		}
	})
}

func TestProjectValidate(t *testing.T) {
	t.Run("valid project passes validation", func(t *testing.T) {
		p := model.Project{
			Document: model.Document{Title: "My Project"},
		}

		err := p.Validate()
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("missing title fails validation", func(t *testing.T) {
		p := model.Project{}

		err := p.Validate()
		if err == nil {
			t.Error("expected error for missing title, got nil")
		}
	})
}

func TestProjectTimespan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		start     string
		end       string
		expected  string
	}{
		{"both dates", "Jan 2023", "Dec 2024", "Jan 2023 — Dec 2024"},
		{"ongoing project", "Mar 2024", "", "Mar 2024 — Present"},
		{"no dates", "", "", ""},
		{"no start ignores end", "", "Dec 2024", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := model.Project{StartDate: tc.start, EndDate: tc.end}

			result := p.Timespan()
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestLetterValidate(t *testing.T) {
	t.Run("valid letter passes validation", func(t *testing.T) {
		l := model.Letter{
			Document: model.Document{Title: "Dear Friend"},
			Date:     "2026-01-01",
		}

		err := l.Validate()
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("missing title fails validation", func(t *testing.T) {
		l := model.Letter{Date: "2026-01-01"}

		err := l.Validate()
		if err == nil {
			t.Error("expected error for missing title, got nil")
		}
	})

	t.Run("missing date fails validation", func(t *testing.T) {
		l := model.Letter{Document: model.Document{Title: "Dear Friend"}}

		err := l.Validate()
		if err == nil {
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

		err := r.Validate()
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("missing title fails validation", func(t *testing.T) {
		r := model.ReadingList{Author: "Alan Donovan"}

		err := r.Validate()
		if err == nil {
			t.Error("expected error for missing title, got nil")
		}
	})

	t.Run("missing author fails validation", func(t *testing.T) {
		r := model.ReadingList{Document: model.Document{Title: "Go Programming"}}

		err := r.Validate()
		if err == nil {
			t.Error("expected error for missing author, got nil")
		}
	})
}
