// Package model provides data models and types for the timterests application.
package model

// Document represents the base structure for various content types.
type Document struct {
	ID       string
	S3Key    string
	Title    string   `yaml:"title"`
	Subtitle string   `yaml:"subtitle"`
	Body     string   `yaml:"body"`
	Tags     []string `yaml:"tags"`
}
