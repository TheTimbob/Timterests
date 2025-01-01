package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Storage struct {
	bucketName string
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
		bucketName: bucketName,
		S3Client:   client,
	}, nil
}

// Gets an object from a bucket and stores it in a local file.
func (storage Storage) DownloadFile(ctx context.Context, objectKey string, fileName string) error {
	result, err := storage.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(storage.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get object %s from bucket %s. No such key exists.\n", objectKey, storage.bucketName)
			err = noKey
		} else {
			log.Printf("Couldn't get object %v:%v. Here's why: %v\n", storage.bucketName, objectKey, err)
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

// Lists the objects in a bucket with a specific prefix.
func (storage Storage) ListObjects(ctx context.Context, prefix string) ([]types.Object, error) {
	var err error
	var output *s3.ListObjectsV2Output
	var objects []types.Object

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(storage.bucketName),
		Prefix: aws.String(prefix),
	}

	objectPaginator := s3.NewListObjectsV2Paginator(storage.S3Client, input)
	for objectPaginator.HasMorePages() {

		output, err = objectPaginator.NextPage(ctx)
		if err != nil {

			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				log.Printf("Bucket %s does not exist.\n", storage.bucketName)
				err = noBucket
			}

			break

		} else {
			objects = append(objects, output.Contents...)
		}
	}
	return objects, err
}

func (storage Storage) Health() map[string]string {

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
