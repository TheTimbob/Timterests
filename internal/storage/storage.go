package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

type Storage struct {
	BucketName string
	S3Client   *s3.Client
}

// Initializes a new Storage instance.
func NewStorage() (*Storage, error) {
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

	return &Storage{
		BucketName: bucketName,
		S3Client:   client,
	}, nil
}

// Lists the objects in a bucket with a specific prefix.
func ListObjects(ctx context.Context, storage Storage, prefix string) ([]types.Object, error) {
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

	// Default sort objects by date, from most current to least current
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(*objects[j].LastModified)
	})

	return objects, err
}

// Gets an object from a bucket and stores it in a local file.
func DownloadFile(ctx context.Context, storage Storage, objectKey string, fileName string) error {

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

	defer func() {
		if err := result.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	file, err := os.Create(fileName)

	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", fileName, err)
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	body, err := io.ReadAll(result.Body)

	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
	}

	_, err = file.Write(body)

	return err
}

// UploadFile uploads a local file to S3
func UploadFileToS3(ctx context.Context, storage Storage, objectKey string, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Couldn't open file %v. Here's why: %v\n", fileName, err)
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	_, err = storage.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(storage.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})

	if err != nil {
		log.Printf("Couldn't upload file %v to %v:%v. Here's why: %v\n", fileName, storage.BucketName, objectKey, err)
	}

	return err
}

// WriteYAMLDocument writes a YAML document to local storage
func WriteYAMLDocument(objectKey string, formData map[string]any) (string, error) {

	fileName := path.Base(objectKey)
	localFilePath := path.Join("s3", fileName)

	if err := os.MkdirAll("s3", 0755); err != nil {
		log.Printf("Couldn't create s3 directory. Here's why: %v\n", err)
		return "", err
	}

	file, err := os.Create(localFilePath)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", localFilePath, err)
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	enc := yaml.NewEncoder(file)
	defer func() {
		if err := enc.Close(); err != nil {
			log.Printf("error closing encoder: %v", err)
		}
	}()

	if err := enc.Encode(formData); err != nil {
		log.Printf("Couldn't encode document to YAML. Here's why: %v\n", err)
		return "", err
	}

	log.Printf("Successfully created document: %s", objectKey)
	return localFilePath, nil
}

func GetPreparedFile(key string, document any, storage Storage) error {
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	// Retrieve file content
	file, err := GetFile(key, localFilePath, storage)
	if err != nil {
		return err
	}

	if err := DecodeFile(file, document); err != nil {
		return err
	}

	if err := BodyToHTML(document); err != nil {
		return err
	}
	return nil
}

// Reads a file from the local file system.
func GetFile(key, localPath string, storage Storage) (*os.File, error) {
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

func DecodeFile(file *os.File, out any) error {
	// Decode the YAML file into a document object
	decoder := yaml.NewDecoder(file)
	// Out should be a pointer to a struct
	if err := decoder.Decode(out); err != nil {
		// Consider using a lower logging level or removing log output
		log.Printf("failed to decode file: %v", err)
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

func GetImageFromS3(storageInstance Storage, imagePath string) (string, error) {
	localImagePath := path.Join("s3", filepath.Base(imagePath))
	err := DownloadFile(context.Background(), storageInstance, imagePath, localImagePath)
	if err != nil {
		log.Printf("Failed to download image: %v", err)
		return localImagePath, err
	}

	localImagePath = path.Join("/", localImagePath)

	return localImagePath, nil
}

// Function to convert body text to HTML
func BodyToHTML(document any) error {
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)

	// Get reflect value of the document
	v := reflect.ValueOf(document)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("document must be a non-nil pointer to a struct")
	}
	v = v.Elem()

	// Set the Body field to the modified HTML content
	bodyField := v.FieldByName("Body")
	body := bodyField.String()

	if err := md.Convert([]byte(body), &buf); err != nil {
		log.Printf("failed to convert body to HTML: %v", err)
		return fmt.Errorf("conversion error: %w", err)
	}

	body = buf.String()

	body = strings.ReplaceAll(body, "<p>", `<p class="content-text">`)
	body = strings.ReplaceAll(body, "<h2>", `<h2 class="category-subtitle">`)
	body = strings.ReplaceAll(body, "<a ", `<a class="hyperlink"`)
	body = strings.ReplaceAll(body, "<li>", `<li class="content-text">- `)

	if bodyField.CanSet() {
		bodyField.SetString(body)
	} else {
		return fmt.Errorf("body field cannot be set in the document struct")
	}

	return nil
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

func SanitizeFilename(filename string) string {
	filename = strings.ToLower(filename)
	filename = strings.ReplaceAll(filename, " ", "-")

	reg := regexp.MustCompile("[^a-z0-9-_]")
	filename = reg.ReplaceAllString(filename, "")

	const maxLength = 50
	if len(filename) > maxLength {
		filename = filename[:maxLength]
	}

	filename = strings.Trim(filename, ".-")

	// Ensure filename is not empty after trimming
	if filename == "" {
		return "unnamed-" + strconv.FormatInt(time.Now().Unix(), 10)
	}
	return filename
}

func Health(storage Storage) map[string]string {
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
