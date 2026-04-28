package web_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"timterests/cmd/web"
	"timterests/internal/auth"
	"timterests/internal/model"
	"timterests/internal/service"

	"github.com/PuerkitoBio/goquery"
)

func TestProjectListRendering(t *testing.T) {
	s := testSetup(t, context.Background())

	t.Run("render project list page", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/projects", nil)
		rec := httptest.NewRecorder()

		web.ProjectsPageHandler(rec, req, *s, "all", "list")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the title of the webpage to be present.
		if doc.Find("title").Length() == 0 {
			t.Error("expected title element to be rendered, but it wasn't")
		}
		// Expect the page name to be set correctly.
		categoryTitle := "Projects"
		if actualPageName := doc.Find("h1.category-title").Text(); actualPageName != categoryTitle {
			t.Errorf("expected page name %q, got %q", categoryTitle, actualPageName)
		}
		// Expect the container element to be present.
		if doc.Find(`[id="projects-container"]`).Length() == 0 {
			t.Error("expected container element to be rendered, but it wasn't")
		}
	})
	t.Run("render project list only", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/projects", nil)
		rec := httptest.NewRecorder()

		// Set the HX-Request header to trigger partial rendering
		req.Header.Set("Hx-Request", "true")

		web.ProjectsPageHandler(rec, req, *s, "all", "list")

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
	t.Run("render projects with a selected tag", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/projects", nil)
		rec := httptest.NewRecorder()

		tag := "Golang"
		web.ProjectsPageHandler(rec, req, *s, tag, "list")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the page-list to be rendered.
		if doc.Find(`ul#page-list`).Length() == 0 {
			t.Error("expected page-list element to be rendered, but it wasn't")
		}
		// Expect at least one project to be rendered.
		if doc.Find(`div.card-container`).Length() == 0 {
			t.Error("expected at least one project to be rendered, but none were")
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

	t.Run("exclude projects that do not have the selected tag", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/projects", nil)
		rec := httptest.NewRecorder()

		// Enter a non-existent tag to get zero results back (filter all projects).
		tag := "non-existent-tag"
		web.ProjectsPageHandler(rec, req, *s, tag, "list")

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}
		// Expect the page-list to be rendered.
		if doc.Find(`ul#page-list`).Length() == 0 {
			t.Error("expected page-list element to be rendered, but it wasn't")
		}
		// Expect no projects to be rendered.
		if doc.Find(`div.card-container`).Length() > 0 {
			t.Error("expected no projects to be rendered, but some were")
		}
	})
}

func TestProjectRendering(t *testing.T) {
	s := testSetup(t, context.Background())

	// Create auth instance for tests (won't be authenticated but prevents nil pointer)
	a := auth.NewAuth("test-session-key-minimum-32-bytes")

	t.Run("render project page", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/project?id=0", nil)
		rec := httptest.NewRecorder()

		web.GetProjectHandler(rec, req, *s, "0", a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		// Expect the title of the webpage to be present
		if doc.Find("title").Length() == 0 {
			t.Error("expected title element to be rendered, but it wasn't")
		}

		// Expect the project content to be present
		if doc.Find("#project-container").Length() == 0 {
			t.Error("expected project container to be rendered, but it wasn't")
		}

		// Expect headers to be present
		if doc.Find("h1").Length() == 0 {
			t.Error("expected h1 element to be rendered, but it wasn't")
		}

		if doc.Find("h2").Length() == 0 {
			t.Error("expected h2 element to be rendered, but it wasn't")
		}
	})

	t.Run("render project display only", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/project?id=0", nil)
		rec := httptest.NewRecorder()

		// Set HTMX header for partial rendering
		req.Header.Set("Hx-Request", "true")

		web.GetProjectHandler(rec, req, *s, "0", a)

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		// Expect the full page title to NOT be rendered
		if doc.Find("title").Length() > 0 {
			t.Error("expected title element to not be rendered, but it was")
		}

		// Expect the project content to be present
		if doc.Find("#project-container").Length() == 0 {
			t.Error("expected project container to be rendered, but it wasn't")
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

func TestGetProjectNotFound(t *testing.T) {
	s := testSetup(t, context.Background())
	a := auth.NewAuth("test-session-key-minimum-32-bytes")

	t.Run("returns 404 for non-existent project ID", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/project?id=non-existent-id", nil)
		rec := httptest.NewRecorder()

		web.GetProjectHandler(rec, req, *s, "non-existent-id", a)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})
}

func TestProjectCardConversion(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("project to card conversion", func(t *testing.T) {
		testProjectPath := "projects/test-project.yaml"

		project, err := service.GetProject(ctx, *s, testProjectPath, 1)
		if err != nil {
			t.Fatalf("failed to get project: %v", err)
		}

		card := web.ProjectCard(*project, 0)

		if card.Title != project.Title {
			t.Errorf("expected card title %q, got %q", project.Title, card.Title)
		}

		if card.Subtitle != project.Subtitle {
			t.Errorf("expected card subtitle %q, got %q", project.Subtitle, card.Subtitle)
		}

		if card.Get != "/project?id=1" {
			t.Errorf("expected card get URL '/project?id=1', got %q", card.Get)
		}

		if card.ImagePath == "" {
			t.Error("expected card image path to be set, but it was empty")
		}

		if card.ImagePath != project.Image {
			t.Errorf("expected card image path %q, got %q", project.Image, card.ImagePath)
		}

		expectedDate := project.Timespan()
		if card.Date != expectedDate {
			t.Errorf("expected card date %q, got %q", expectedDate, card.Date)
		}
	})
}

func TestProjectRepositoryRendering(t *testing.T) {
	dc := model.DisplayContent{ID: "0", S3Key: "projects/test.yaml", Body: ""}

	renderDisplay := func(t *testing.T, repository string) *goquery.Document {
		t.Helper()

		var buf bytes.Buffer

		err := web.ProjectDisplay(dc, repository, "", false).Render(context.Background(), &buf)
		if err != nil {
			t.Fatalf("failed to render ProjectDisplay: %v", err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(buf.String()))
		if err != nil {
			t.Fatalf("failed to parse rendered HTML: %v", err)
		}

		return doc
	}

	t.Run("hides repository link when empty", func(t *testing.T) {
		doc := renderDisplay(t, "")

		if doc.Find("#project-container a").Length() > 0 {
			t.Error("expected no repository link when repository is empty, but one was rendered")
		}
	})

	t.Run("hides repository link when Private", func(t *testing.T) {
		doc := renderDisplay(t, "Private")

		if doc.Find("#project-container a").Length() > 0 {
			t.Error("expected no repository link when repository is 'Private', but one was rendered")
		}
	})

	t.Run("renders repository link when URL is set", func(t *testing.T) {
		repo := "https://github.com/test/repo"
		doc := renderDisplay(t, repo)

		link := doc.Find("#project-container a")
		if link.Length() == 0 {
			t.Error("expected repository link to be rendered, but it was not")
		}

		if href, _ := link.Attr("href"); href != repo {
			t.Errorf("expected href %q, got %q", repo, href)
		}
	})
}
