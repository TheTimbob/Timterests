package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Storage provides S3 storage operations with bucket and client configuration.
type Storage struct {
	BucketName string
	S3Client   *s3.Client
}

// NewS3Storage initializes a new Storage instance with AWS S3 configuration.
func NewS3Storage() (*Storage, error) {
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	region := os.Getenv("AWS_REGION")

	if bucketName == "" || region == "" {
		return nil, errors.New("BUCKET_NAME or AWS_REGION is not set in the environment variables")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &Storage{
		BucketName: bucketName,
		S3Client:   client,
	}, nil
}

// ListS3Objects lists the objects in a bucket with a specific prefix.
func (s *Storage) ListS3Objects(ctx context.Context, prefix string) ([]types.Object, error) {
	var (
		err     error
		output  *s3.ListObjectsV2Output
		objects []types.Object
	)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.BucketName),
		Prefix: aws.String(prefix),
	}

	objectPaginator := s3.NewListObjectsV2Paginator(s.S3Client, input)
	for objectPaginator.HasMorePages() {
		output, err = objectPaginator.NextPage(ctx)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				log.Printf("Bucket %s does not exist.\n", s.BucketName)

				err = noBucket
			}

			break
		}

		objects = append(objects, output.Contents...)
	}

	// Default sort objects by date, from most current to least current
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(*objects[j].LastModified)
	})

	return objects, err
}

// DownloadS3File gets an object from a bucket and stores it in a local file.
func (s *Storage) DownloadS3File(ctx context.Context, objectKey string, fileName string) error {
	result, err := s.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get object %s from bucket %s. No such key exists.\n", objectKey, s.BucketName)

			err = noKey
		} else {
			log.Printf("Couldn't get object %v:%v. Here's why: %v\n", s.BucketName, objectKey, err)
		}

		return err
	}

	defer func() {
		err := result.Body.Close()
		if err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	// #nosec G304 -- fileName is constructed from S3 object keys which are controlled by the application
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", fileName, err)

		return fmt.Errorf("failed to create file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
	}

	_, err = file.Write(body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// UploadFileToS3 uploads a local file to S3.
func (s *Storage) UploadFileToS3(ctx context.Context, objectKey string, fileName string) error {
	// #nosec G304 -- fileName is constructed from sanitized user input and restricted to s3/ directory
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Couldn't open file %v. Here's why: %v\n", fileName, err)

		return fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	_, err = s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		log.Printf("Couldn't upload file %v to %v:%v. Here's why: %v\n", fileName, s.BucketName, objectKey, err)

		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// GetPreparedFile retrieves a file from S3, decodes it, and converts markdown to HTML.
func (s *Storage) GetPreparedFile(ctx context.Context, key string, document any) error {
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	// Retrieve file content
	file, err := s.GetFile(ctx, key, localFilePath)
	if err != nil {
		return err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	err = DecodeFile(file, document)
	if err != nil {
		return err
	}

	err = BodyToHTML(document)
	if err != nil {
		return err
	}

	return nil
}

// GetFile downloads a file from S3 and opens it from the local file system.
func (s *Storage) GetFile(ctx context.Context, key, localPath string) (*os.File, error) {
	var file *os.File

	// Download the file
	err := s.DownloadS3File(ctx, key, localPath)
	if err != nil {
		return file, err
	}

	// Open the downloaded file
	// #nosec G304 -- localPath is constructed from controlled S3 object keys
	file, err = os.Open(localPath)
	if err != nil {
		return file, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// GetImageFromS3 downloads an image from S3 and returns its local path.
func (s *Storage) GetImageFromS3(ctx context.Context, imagePath string) (string, error) {
	localImagePath := path.Join("s3", filepath.Base(imagePath))

	err := s.DownloadS3File(ctx, imagePath, localImagePath)
	if err != nil {
		log.Printf("Failed to download image: %v", err)

		return localImagePath, err
	}

	localImagePath = path.Join("/", localImagePath)

	return localImagePath, nil
}

// Health checks the S3 connection status and returns health information.
func (s *Storage) Health() map[string]string {
	health := make(map[string]string)

	_, err := s.S3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		health["status"] = "down"
		health["message"] = fmt.Sprintf("Failed to list buckets: %v", err)
		log.Fatalf("S3 connection down: %v", err) // Log the error and terminate the program

		return health
	}

	health["status"] = "up"
	health["message"] = "S3 storage is up and running."

	return health
}
