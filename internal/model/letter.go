package model

import "errors"

// Letter represents a personal letter with date and occasion.
type Letter struct {
	Document `yaml:",inline"`

	Date     string `yaml:"date"`
	Occasion string `yaml:"occasion"`
}

// Validate checks that the Letter has the required fields populated.
func (l *Letter) Validate() error {
	if l.Title == "" {
		return errors.New("letter title is required")
	}

	if l.Date == "" {
		return errors.New("letter date is required")
	}

	return nil
}
