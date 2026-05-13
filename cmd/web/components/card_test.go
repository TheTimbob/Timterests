package components_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"timterests/cmd/web/components"
)

func render(t *testing.T, c interface{ Render(ctx context.Context, w io.Writer) error }) string {
	t.Helper()

	var buf bytes.Buffer

	err := c.Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	return buf.String()
}

func TestLargeCard(t *testing.T) {
	t.Parallel()

	t.Run("renders all fields", func(t *testing.T) {
		t.Parallel()

		c := components.Card{
			Title:     "Test Title",
			Subtitle:  "Test Sub",
			Date:      "2026-01-15",
			Preview:   "Some preview text.",
			ImagePath: "/images/test.png",
			Get:       "/article?id=1",
			Tags:      []string{"go", "testing"},
			Index:     2,
		}

		html := render(t, c.LargeCard())

		for _, want := range []string{
			"Test Title",
			"Test Sub",
			"2026-01-15",
			"Some preview text.",
			"/images/test.png",
			"/article?id=1",
			"go",
			"testing",
			"card-container",
			"animation-delay: 0.2s",
		} {
			if !strings.Contains(html, want) {
				t.Errorf("LargeCard missing %q", want)
			}
		}
	})

	t.Run("omits image when ImagePath is empty", func(t *testing.T) {
		t.Parallel()

		c := components.Card{Title: "No Image", Get: "/x"}
		html := render(t, c.LargeCard())

		if strings.Contains(html, "card-image") {
			t.Error("expected no card-image when ImagePath is empty")
		}
	})

	t.Run("omits date when empty", func(t *testing.T) {
		t.Parallel()

		c := components.Card{Title: "No Date", Get: "/x"}
		html := render(t, c.LargeCard())

		if strings.Contains(html, "card-date") {
			t.Error("expected no card-date when Date is empty")
		}
	})
}

func TestMiniCard(t *testing.T) {
	t.Parallel()

	t.Run("renders title and tags", func(t *testing.T) {
		t.Parallel()

		c := components.Card{
			Title: "Mini Title",
			Get:   "/mini",
			Tags:  []string{"rust"},
			Index: 0,
		}

		html := render(t, c.MiniCard())

		for _, want := range []string{
			"Mini Title",
			"mini-card-container",
			"rust",
			"animation-delay: 0.0s",
		} {
			if !strings.Contains(html, want) {
				t.Errorf("MiniCard missing %q", want)
			}
		}
	})

	t.Run("renders image when present", func(t *testing.T) {
		t.Parallel()

		c := components.Card{Title: "With Img", Get: "/x", ImagePath: "/img/cover.jpg"}
		html := render(t, c.MiniCard())

		if !strings.Contains(html, "card-image") {
			t.Error("expected card-image in MiniCard with ImagePath")
		}
	})
}

func TestLinkCard(t *testing.T) {
	t.Parallel()

	t.Run("with date shows date dash title", func(t *testing.T) {
		t.Parallel()

		c := components.Card{Title: "Link Title", Date: "2026-03-01", Get: "/link"}
		html := render(t, c.LinkCard())

		if !strings.Contains(html, "2026-03-01") {
			t.Error("expected date in LinkCard")
		}

		if !strings.Contains(html, "Link Title") {
			t.Error("expected title in LinkCard")
		}

		if !strings.Contains(html, "link-item") {
			t.Error("expected link-item class")
		}
	})

	t.Run("without date shows only title", func(t *testing.T) {
		t.Parallel()

		c := components.Card{Title: "No Date Link", Get: "/link2"}
		html := render(t, c.LinkCard())

		if !strings.Contains(html, "No Date Link") {
			t.Error("expected title")
		}

		if strings.Contains(html, " - ") {
			t.Error("expected no dash separator without date")
		}
	})
}

func TestCardTitleContainer(t *testing.T) {
	t.Parallel()

	html := render(t, components.CardTitleContainer("Main Title", "Sub Title"))

	for _, want := range []string{
		"card-title-container",
		"Main Title",
		"Sub Title",
		"card-title",
		"card-subtitle",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("CardTitleContainer missing %q", want)
		}
	}
}

func TestAnimationDelay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		index int
		want  string
	}{
		{0, "0.0s"},
		{1, "0.1s"},
		{5, "0.5s"},
		{10, "1.0s"},
	}

	for _, tc := range tests {
		c := components.Card{Title: "T", Get: "/x", Index: tc.index}
		html := render(t, c.LargeCard())

		if !strings.Contains(html, tc.want) {
			t.Errorf("Index %d: expected delay %q in HTML", tc.index, tc.want)
		}
	}
}
