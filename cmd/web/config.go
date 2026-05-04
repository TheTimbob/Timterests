package web

import "os"

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
}

func site() SiteConfig {
	return SiteConfig{
		Name:           envOr("SITE_NAME", "Timterests"),
		Subtitle:       envOr("SITE_SUBTITLE", "Tim's interests"),
		AuthorName:     envOr("AUTHOR_NAME", "Tim Scott"),
		URL:            envOr("SITE_URL", "https://timterests.com"),
		Description:    envOr("SITE_DESCRIPTION", "Tim Scott's personal site — articles, projects, and a curated reading list."),
		RepoURL:        envOr("REPO_URL", "https://github.com/TheTimbob/timterests"),
		FontAwesomeKit: envOr("FONTAWESOME_KIT_ID", "3453ab8a44"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
