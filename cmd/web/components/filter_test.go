package components_test

import (
	"strings"
	"testing"

	"timterests/cmd/web/components"
)

func TestFilterTags(t *testing.T) {
	t.Parallel()

	t.Run("renders all tag options plus All", func(t *testing.T) {
		t.Parallel()

		html := render(t, components.FilterTags("/articles", []string{"go", "rust", "python"}))

		for _, want := range []string{
			"filter-select",
			`value="all"`,
			"All",
			`value="go"`,
			`value="rust"`,
			`value="python"`,
			`hx-get="/articles"`,
			`name="tag"`,
		} {
			if !strings.Contains(html, want) {
				t.Errorf("FilterTags missing %q", want)
			}
		}
	})

	t.Run("empty tags renders only All option", func(t *testing.T) {
		t.Parallel()

		html := render(t, components.FilterTags("/x", nil))

		if !strings.Contains(html, "All") {
			t.Error("expected All option")
		}

		if count := strings.Count(html, "<option"); count != 1 {
			t.Errorf("expected 1 option element, got %d", count)
		}
	})
}

func TestFilterDesign(t *testing.T) {
	t.Parallel()

	t.Run("list design is active by default", func(t *testing.T) {
		t.Parallel()

		html := render(t, components.FilterDesign("/articles", "list"))

		if !strings.Contains(html, "view-options") {
			t.Error("expected view-options container")
		}

		buttons := strings.Count(html, "view-btn")
		if buttons < 3 {
			t.Errorf("expected at least 3 view-btn references, got %d", buttons)
		}
	})

	t.Run("empty design defaults to list active", func(t *testing.T) {
		t.Parallel()

		html := render(t, components.FilterDesign("/articles", ""))

		if !strings.Contains(html, `"design": "list"`) {
			t.Error("expected list design hx-vals")
		}

		if !strings.Contains(html, `"design": "grid"`) {
			t.Error("expected grid design hx-vals")
		}
	})

	t.Run("grid design renders grid button active", func(t *testing.T) {
		t.Parallel()

		html := render(t, components.FilterDesign("/projects", "grid"))

		if !strings.Contains(html, `hx-get="/projects"`) {
			t.Error("expected hx-get for projects")
		}
	})
}
