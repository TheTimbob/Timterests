package service_test

import (
	"context"
	"slices"
	"testing"
	"timterests/internal/service"
)

func TestListProjects(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("returns all projects when tag is empty", func(t *testing.T) {
		projects, err := service.ListProjects(ctx, *s, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(projects) < 1 {
			t.Errorf("expected at least one project, got %d", len(projects))
		}
	})

	t.Run("returns all projects when tag is 'all'", func(t *testing.T) {
		projects, err := service.ListProjects(ctx, *s, "all")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(projects) < 1 {
			t.Errorf("expected at least one project, got %d", len(projects))
		}
	})

	t.Run("filters projects by tag", func(t *testing.T) {
		projects, err := service.ListProjects(ctx, *s, "Golang")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(projects) < 1 {
			t.Errorf("expected at least one project with tag 'Golang', got %d", len(projects))
		}

		for _, p := range projects {
			if !slices.Contains(p.Tags, "Golang") {
				t.Errorf("project %q does not have tag 'Golang'", p.Title)
			}
		}
	})

	t.Run("returns empty slice for non-existent tag", func(t *testing.T) {
		projects, err := service.ListProjects(ctx, *s, "does-not-exist")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(projects) != 0 {
			t.Errorf("expected zero projects for non-existent tag, got %d", len(projects))
		}
	})
}

func TestGetProject(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("retrieves project by key and id", func(t *testing.T) {
		project, err := service.GetProject(ctx, *s, "projects/test-project.md", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if project.ID != "0" {
			t.Errorf("expected ID '0', got %q", project.ID)
		}

		if project.Title != "Test Project" {
			t.Errorf("expected title 'Test Project', got %q", project.Title)
		}
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		_, err := service.GetProject(ctx, *s, "projects/does-not-exist.md", 0)
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})
}

func TestGetFeaturedProject(t *testing.T) {
	ctx := context.Background()
	s := testSetup(t, ctx)

	t.Run("returns project matching title", func(t *testing.T) {
		project, err := service.GetFeaturedProject(ctx, *s, "Test Project")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if project.Title != "Test Project" {
			t.Errorf("expected title 'Test Project', got %q", project.Title)
		}
	})

	t.Run("returns error when title does not match any project", func(t *testing.T) {
		_, err := service.GetFeaturedProject(ctx, *s, "Does Not Exist")
		if err == nil {
			t.Error("expected error for non-existent featured project, got nil")
		}
	})
}
