package types

type Document struct {
	ID       string
	Title    string   `yaml:"title"`
	Subtitle string   `yaml:"subtitle"`
	Body     string   `yaml:"body"`
	Tags     []string `yaml:"tags"`
}

// DocumentItem interface that all document types must implement
type DocumentItem interface {
	GetID() string
	GetTitle() string
	GetSubtitle() string
	GetBody() string
	GetTags() []string
	SetID(id string)
	SetBody(body string)
}

// Implement DocumentItem for Document
func (d Document) GetID() string        { return d.ID }
func (d Document) GetTitle() string     { return d.Title }
func (d Document) GetSubtitle() string  { return d.Subtitle }
func (d Document) GetBody() string      { return d.Body }
func (d Document) GetTags() []string    { return d.Tags }
func (d *Document) SetID(id string)     { d.ID = id }
func (d *Document) SetBody(body string) { d.Body = body }
