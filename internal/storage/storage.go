package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"timterests/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

// Initializes a new models.Storage instance.
func NewStorage() (*models.Storage, error) {
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	region := os.Getenv("AWS_REGION")

	if bucketName == "" || region == "" {
		return nil, fmt.Errorf("BUCKET_NAME or AWS_REGION is not set in the environment variables")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg)

	return &models.Storage{
		BucketName: bucketName,
		S3Client:   client,
	}, nil
}

// Lists the objects in a bucket with a specific prefix.
func ListObjects(ctx context.Context, storage models.Storage, prefix string) ([]types.Object, error) {
	var err error
	var output *s3.ListObjectsV2Output
	var objects []types.Object

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(storage.BucketName),
		Prefix: aws.String(prefix),
	}

	objectPaginator := s3.NewListObjectsV2Paginator(storage.S3Client, input)
	for objectPaginator.HasMorePages() {

		output, err = objectPaginator.NextPage(ctx)
		if err != nil {

			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				log.Printf("Bucket %s does not exist.\n", storage.BucketName)
				err = noBucket
			}

			break

		} else {
			objects = append(objects, output.Contents...)
		}
	}

	// Sort objects by date from most current to least current
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(*objects[j].LastModified)
	})

	return objects, err
}

// Gets an object from a bucket and stores it in a local file.
func DownloadFile(ctx context.Context, storage models.Storage, objectKey string, fileName string) error {

	if _, err := os.Stat(fileName); err == nil {
		return nil
	}

	result, err := storage.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(storage.BucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get object %s from bucket %s. No such key exists.\n", objectKey, storage.BucketName)
			err = noKey
		} else {
			log.Printf("Couldn't get object %v:%v. Here's why: %v\n", storage.BucketName, objectKey, err)
		}
		return err
	}

	defer result.Body.Close()

	file, err := os.Create(fileName)

	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", fileName, err)
		return err
	}

	defer file.Close()

	body, err := io.ReadAll(result.Body)

	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
	}

	_, err = file.Write(body)

	return err
}

// Reads a file from the local file system.
func GetFile(key, localPath string, storage models.Storage) (*os.File, error) {
	var file *os.File

	// Download the file
	err := DownloadFile(context.Background(), storage, key, localPath)
	if err != nil {
		return file, err
	}

	// Open the downloaded file
	file, err = os.Open(localPath)
	if err != nil {
		return file, err
	}

	return file, nil
}

func DecodeFile(file *os.File, out interface{}) error {

	// Decode the yaml file into a document object
	decoder := yaml.NewDecoder(file)

	// Out should be a pointer to a struct
	if err := decoder.Decode(out); err != nil {
		log.Printf("Failed to decode file: %v", err)
		return err
	}
	return nil
}

// Function to convert body text to HTML
func BodyToHTML(str string) (string, error) {
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)
	err := md.Convert([]byte(str), &buf)

	str = buf.String()

	str = strings.ReplaceAll(str, "<p>", `<p class="content-text">`)
	str = strings.ReplaceAll(str, "<h2>", `<h2 class="category-subtitle">`)
	str = strings.ReplaceAll(str, "<a ", `<a class="hyperlink"`)
    str = strings.ReplaceAll(str, "<li>", `<li class="content-text">- `)

	return str, err
}

func GetTags(v reflect.Value, tags []string) []string {

	field := v.FieldByName("Tags")

	// If the Tags field is not directly on the struct, check the embedded Document
	if !field.IsValid() {
		embeddedDoc := v.FieldByName("Document")
		if embeddedDoc.IsValid() {
			field = embeddedDoc.FieldByName("Tags")
		}
	}

	// Create a list of tags
	for i := 0; i < field.Len(); i++ {
		tag := field.Index(i).String()
		if !slices.Contains(tags, tag) {
			tags = append(tags, tag)
		}
	}

	return tags
}

func RemoveHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

func Health(storage models.Storage) map[string]string {
	health := make(map[string]string)

	_, err := storage.S3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		health["status"] = "down"
		health["message"] = fmt.Sprintf("Failed to list buckets: %v", err)
		log.Fatalf("S3 connection down: %v", err) // Log the error and terminate the program
	} else {
		health["status"] = "up"
		health["message"] = "S3 storage is up and running."
	}

	return health
}
