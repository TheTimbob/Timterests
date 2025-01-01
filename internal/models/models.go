package models

type Article struct {
	ID       string
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	Date     string `yaml:"date"`
	Body     string `yaml:"body"`
}
