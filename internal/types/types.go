package types

type Document struct {
	ID       string
	Title    string   `yaml:"title"`
	Subtitle string   `yaml:"subtitle"`
	Body     string   `yaml:"body"`
	Tags     []string `yaml:"tags"`
}
