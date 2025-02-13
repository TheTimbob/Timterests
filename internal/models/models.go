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
}

type About struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	Body     string `yaml:"body"`
    Tags     []string `yaml:"tags"`
}

type Project struct {
	Document   `yaml:",inline"`
	Repository string `yaml:"repository"`
	Image      string `yaml:"image-path"`
    Tags     []string `yaml:"tags"`
}

type Article struct {
	Document `yaml:",inline"`
	Date     string `yaml:"date"`
    Tags     []string `yaml:"tags"`
}

type ReadingList struct {
	Document `yaml:",inline"`
	Image    string `yaml:"image-path"`
    Tags     []string `yaml:"tags"`
}

type Letter struct {
	Document `yaml:",inline"`
    Date     string `yaml:"date"`
    Occasion string `yaml:"occasion"`
    Password string `yaml:"password"`
}