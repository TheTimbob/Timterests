package model

import "errors"

// Article represents a blog article with metadata and content.
type Article struct {
	Document `yaml:",inline"`

	Date string `yaml:"date"`
}

// Validate checks that the Article has the required fields populated.
func (a *Article) Validate() error {
	if a.Title == "" {
		return errors.New("article title is required")
	}

	if a.Date == "" {
		return errors.New("article date is required")
	}

	return nil
}
