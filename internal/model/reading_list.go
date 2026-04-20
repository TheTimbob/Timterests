package model

import "errors"

// ReadingList represents a book that appears in the reading list.
type ReadingList struct {
	Document `yaml:",inline"`

	Image     string `yaml:"imagePath"`
	Author    string `yaml:"author"`
	Published string `yaml:"published"`
	ISBN      string `yaml:"isbn"`
	Website   string `yaml:"website"`
}

// Validate checks that the ReadingList entry has the required fields populated.
func (r *ReadingList) Validate() error {
	if r.Title == "" {
		return errors.New("book title is required")
	}

	if r.Author == "" {
		return errors.New("book author is required")
	}

	return nil
}
