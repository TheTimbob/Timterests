package web

import (
	"testing"
)

func TestSiteConfigDefaults(t *testing.T) {
	t.Setenv("SITE_NAME", "")
	t.Setenv("AUTHOR_NAME", "")
	t.Setenv("SITE_URL", "")

	cfg := site()

	if cfg.Name != "Timterests" {
		t.Errorf("expected default Name %q, got %q", "Timterests", cfg.Name)
	}

	if cfg.AuthorName != "Tim Scott" {
		t.Errorf("expected default AuthorName %q, got %q", "Tim Scott", cfg.AuthorName)
	}

	if cfg.URL != "https://timterests.com" {
		t.Errorf("expected default URL %q, got %q", "https://timterests.com", cfg.URL)
	}
}

func TestSiteConfigFromEnv(t *testing.T) {
	t.Setenv("SITE_NAME", "TestSite")
	t.Setenv("AUTHOR_NAME", "Jane Doe")
	t.Setenv("SITE_URL", "https://example.com")
	t.Setenv("SITE_SUBTITLE", "A test site")
	t.Setenv("REPO_URL", "https://github.com/test/repo")

	cfg := site()

	if cfg.Name != "TestSite" {
		t.Errorf("expected Name %q, got %q", "TestSite", cfg.Name)
	}

	if cfg.AuthorName != "Jane Doe" {
		t.Errorf("expected AuthorName %q, got %q", "Jane Doe", cfg.AuthorName)
	}

	if cfg.Subtitle != "A test site" {
		t.Errorf("expected Subtitle %q, got %q", "A test site", cfg.Subtitle)
	}

	if cfg.RepoURL != "https://github.com/test/repo" {
		t.Errorf("expected RepoURL %q, got %q", "https://github.com/test/repo", cfg.RepoURL)
	}
}
