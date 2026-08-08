package model

// Letter represents a personal letter with date and occasion.
type Letter struct {
	Document `yaml:",inline"`

	Date     string `validate:"required" yaml:"date"`
	Occasion string `yaml:"occasion"`
}

// Validate checks that the Letter has the required fields populated.
func (l *Letter) Validate() error {
	return ValidateRequired(l)
}
