package storage

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"timterests/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

// Reads a file from the local file system and decodes it into a Document object.
func ReadFile(key, localFilePath string, storageInstance models.Storage) (models.Document, error) {
	var document models.Document

	// Download the file
	err := DownloadFile(context.Background(), storageInstance, key, localFilePath)
	if err != nil {
		log.Println("Failed to download file: ", err)
		return document, err
	}

	// Open the downloaded file
	file, err := os.Open(localFilePath)
	if err != nil {
		log.Println("Failed to open file: ", err)
		return document, err
	}
	defer file.Close()

	// Decode the yaml file into a document object
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&document); err != nil {
		log.Println("Failed to decode file: ", err)
		return document, err
	}
	return document, nil
}

// Converts raw text into HTML paragraphs
func ConvertTextToParagraphs(text string) string {
    paragraphs := strings.Split(text, "\n\n") // Split by double newline for paragraphs
    var htmlContent string
    htmlTagRegex := regexp.MustCompile(`</?[a-z][\s\S]*>`)

    for _, paragraph := range paragraphs {
        if htmlTagRegex.MatchString(paragraph) {
            // If the paragraph contains HTML tags, add it as is
            htmlContent += paragraph
        } else {
            // Escape any special HTML characters to prevent injection
            htmlContent += "<p class='content-text'>" + html.EscapeString(paragraph) + "</p>"
        }
    }

    return htmlContent
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
