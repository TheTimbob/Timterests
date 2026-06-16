package web

import (
	"os"
	"strings"
)

// SiteConfig holds site identity values read from environment variables.
// Defaults match the original hardcoded Timterests values.
type SiteConfig struct {
	Name           string // SITE_NAME
	Subtitle       string // SITE_SUBTITLE
	AuthorName     string // AUTHOR_NAME
	URL            string // SITE_URL
	Description    string // SITE_DESCRIPTION
	RepoURL        string // REPO_URL
	FontAwesomeKit string // FONTAWESOME_KIT_ID
	hiddenPages    map[string]struct{}
}

// IsHidden reports whether a named page (e.g. "articles", "projects", "reading-list")
// has been disabled via the HIDDEN_PAGES environment variable.
func (s SiteConfig) IsHidden(page string) bool {
	_, ok := s.hiddenPages[page]
	return ok
}

// Site returns the current site configuration from environment variables.
func Site() SiteConfig {
	hidden := make(map[string]struct{})

	if v := os.Getenv("HIDDEN_PAGES"); v != "" {
		for _, p := range strings.Split(v, ",") {
			if p = strings.TrimSpace(p); p != "" {
				hidden[p] = struct{}{}
			}
		}
	}

	return SiteConfig{
		Name:           envOr("SITE_NAME", "Timterests"),
		Subtitle:       envOr("SITE_SUBTITLE", "Tim's interests"),
		AuthorName:     envOr("AUTHOR_NAME", "Tim Scott"),
		URL:            envOr("SITE_URL", "https://timterests.com"),
		Description: envOr("SITE_DESCRIPTION",
			"Tim Scott's personal site — articles, projects, and a curated reading list."),
		RepoURL:        envOr("REPO_URL", "https://github.com/TheTimbob/timterests"),
		FontAwesomeKit: envOr("FONTAWESOME_KIT_ID", "3453ab8a44"),
		hiddenPages:    hidden,
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
