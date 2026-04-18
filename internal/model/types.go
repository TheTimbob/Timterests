// Package model provides data models and types for the timterests application.
package model

// Document represents the base structure for various content types.
type Document struct {
	ID       string
	S3Key    string
	Title    string   `yaml:"title"`
	Subtitle string   `yaml:"subtitle"`
	Preview  string   `yaml:"preview"`
	Tags     []string `yaml:"tags"`
}

// SetMeta sets the ID and S3Key fields on the document.
func (d *Document) SetMeta(id, key string) {
	d.ID = id
	d.S3Key = key
}

// DisplayContent holds the minimal document identity (ID, S3Key) paired with its body content.
// This is used for rendering documents, where only the ID/S3Key are needed for navigation/links.
type DisplayContent struct {
	ID    string
	S3Key string
	Body  string
}

// Content holds a document with its body for editor forms (raw markdown).
type Content[T any] struct {
	Doc  T
	Body string
}
