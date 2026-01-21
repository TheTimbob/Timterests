package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"timterests/cmd/web"
)

func TestAboutHandler(t *testing.T) {
	s := testSetup(t, context.Background())

	t.Run("render about page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/about", nil)
		rec := httptest.NewRecorder()

		web.AboutHandler(rec, req, *s)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if rec.Body.Len() == 0 {
			t.Error("expected non-empty response body")
		}
	})

	t.Run("render about tabs", func(t *testing.T) {
		t.Parallel()

		tabs := []string{"bio", "education", "work", "skills"}

		for _, tab := range tabs {
			req := httptest.NewRequest(http.MethodGet, "/about?tab="+tab, nil)
			rec := httptest.NewRecorder()

			web.AboutHandler(rec, req, *s)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200 for tab %s, got %d", tab, rec.Code)
			}

			if rec.Body.Len() == 0 {
				t.Errorf("expected non-empty response body for tab %s", tab)
			}
		}
	})
}

func TestAboutContent(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("load about content", func(t *testing.T) {
		var about web.About

		prefix := "about/"

		aboutFile, err := s.ListObjects(ctx, prefix)
		if err != nil {
			t.Fatalf("failed to list about files: %v", err)
		}

		if len(aboutFile) == 0 {
			t.Fatal("no about files found")
		}

		key := *aboutFile[0].Key

		err = s.GetPreparedFile(ctx, key, &about)
		if err != nil {
			t.Fatalf("failed to get about content: %v", err)
		}

		if about.Title != "About" {
			t.Error("expected title to be 'About'")
		}

		if len(about.Experience) > 0 && about.Experience[0].Role != "Software Engineer" {
			t.Error("expected experience role to be 'Software Engineer'")
		}

		if len(about.Education) > 0 && about.Education[0].Degree != "Bachelors degree" {
			t.Error("expected education degree to be 'Bachelors degree'")
		}

		if len(about.Skills) > 0 && about.Skills[0].Name != "Skill Type One" {
			t.Error("expected skill name to be 'Skill Type One'")
		}
	})
}
