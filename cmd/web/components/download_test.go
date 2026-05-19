package components_test

import (
	"strings"
	"testing"

	"timterests/cmd/web/components"
)

func TestDownloadDocumentButton(t *testing.T) {
	t.Parallel()

	t.Run("renders button for admin", func(t *testing.T) {
		t.Parallel()

		html := render(t, components.DownloadDocumentButton("articles/doc.yaml", true))

		for _, want := range []string{
			`action="/download"`,
			`name="key"`,
			`value="articles/doc.yaml"`,
			"Download",
		} {
			if !strings.Contains(html, want) {
				t.Errorf("DownloadDocumentButton missing %q", want)
			}
		}
	})

	t.Run("renders nothing for non-admin", func(t *testing.T) {
		t.Parallel()

		html := render(t, components.DownloadDocumentButton("articles/doc.yaml", false))

		if strings.Contains(html, "Download") {
			t.Error("expected no content for non-admin")
		}
	})
}

func TestDownloadNewDocumentButton(t *testing.T) {
	t.Parallel()

	t.Run("renders button for admin", func(t *testing.T) {
		t.Parallel()

		html := render(t, components.DownloadNewDocumentButton(true))

		if !strings.Contains(html, `formaction="/download/new"`) {
			t.Error("expected formaction for download/new")
		}
	})

	t.Run("renders nothing for non-admin", func(t *testing.T) {
		t.Parallel()

		html := render(t, components.DownloadNewDocumentButton(false))

		if strings.Contains(html, "Download") {
			t.Error("expected no content for non-admin")
		}
	})
}
