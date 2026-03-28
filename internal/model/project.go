package model

import "errors"

// Project represents a personal software project.
type Project struct {
	Document `yaml:",inline"`

	Repository string `yaml:"repository"`
	Image      string `yaml:"imagePath"`
}

// Validate checks that the Project has the required fields populated.
func (p *Project) Validate() error {
	if p.Title == "" {
		return errors.New("project title is required")
	}

	return nil
}
