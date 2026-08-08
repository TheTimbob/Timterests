package model

// Project represents a personal software project.
type Project struct {
	Document `yaml:",inline"`

	Repository string `yaml:"repository"`
	Image      string `yaml:"imagePath"`
	StartDate  string `yaml:"startDate"`
	EndDate    string `yaml:"endDate"`
}

// Timespan returns a formatted date range for display.
func (p *Project) Timespan() string {
	if p.StartDate == "" {
		return ""
	}

	if p.EndDate == "" {
		return p.StartDate + " — Present"
	}

	return p.StartDate + " — " + p.EndDate
}

// Validate checks that the Project has the required fields populated.
func (p *Project) Validate() error {
	return ValidateRequired(p)
}
