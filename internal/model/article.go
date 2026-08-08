package model

// Article represents a blog article with metadata and content.
type Article struct {
	Document `yaml:",inline"`

	Date string `validate:"required" yaml:"date"`
}

// Validate checks that the Article has the required fields populated.
func (a *Article) Validate() error {
	return ValidateRequired(a)
}
