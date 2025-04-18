package models

import "github.com/aws/aws-sdk-go-v2/service/s3"

type Storage struct {
	BucketName string
	S3Client   *s3.Client
}

type Document struct {
	ID       string
	Title    string   `yaml:"title"`
	Subtitle string   `yaml:"subtitle"`
	Body     string   `yaml:"body"`
	Tags     []string `yaml:"tags"`
}

type About struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	Body     string `yaml:"body"`
}

type Project struct {
	Document   `yaml:",inline"`
	Repository string `yaml:"repository"`
	Image      string `yaml:"image-path"`
}

type Article struct {
	Document `yaml:",inline"`
	Date     string `yaml:"date"`
}

type ReadingList struct {
	Document  `yaml:",inline"`
	Image     string `yaml:"image-path"`
	Author    string `yaml:"author"`
	Published string `yaml:"published"`
	ISBN      string `yaml:"isbn"`
	Website   string `yaml:"website"`
	Status    string `yaml:"status"`
}

type Letter struct {
	Document `yaml:",inline"`
	Date     string `yaml:"date"`
	Occasion string `yaml:"occasion"`
}
