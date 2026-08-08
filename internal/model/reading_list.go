package model

// ReadingList represents a book that appears in the reading list.
type ReadingList struct {
	Document `yaml:",inline"`

	Image     string `yaml:"imagePath"`
	Author    string `validate:"required" yaml:"author"`
	Published string `yaml:"published"`
	ISBN      string `yaml:"isbn"`
	Website   string `yaml:"website"`
}

// Validate checks that the ReadingList entry has the required fields populated.
func (r *ReadingList) Validate() error {
	return ValidateRequired(r)
}
