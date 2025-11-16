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

type Storage struct {
	BucketName string
	S3Client   *s3.Client
}

// Initializes a new Storage instance.
func NewS3Storage() (*Storage, error) {
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
func (s *Storage) ListS3Objects(ctx context.Context, prefix string) ([]types.Object, error) {
	var err error
	var output *s3.ListObjectsV2Output
	var objects []types.Object

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
func (s *Storage) UploadFileToS3(ctx context.Context, objectKey string, fileName string) error {
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

	_, err = s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})

	if err != nil {
		log.Printf("Couldn't upload file %v to %v:%v. Here's why: %v\n", fileName, s.BucketName, objectKey, err)
	}

	return err
}

func (s *Storage) GetPreparedFile(key string, document any) error {
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	// Retrieve file content
	file, err := s.GetFile(key, localFilePath)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	if err := DecodeFile(file, document); err != nil {
		return err
	}

	if err := BodyToHTML(document); err != nil {
		return err
	}
	return nil
}

// Reads a file from the local file system.
func (s *Storage) GetFile(key, localPath string) (*os.File, error) {
	var file *os.File

	// Download the file
	err := s.DownloadS3File(context.Background(), key, localPath)
	if err != nil {
		return file, err
	}

	// Open the downloaded file
	file, err = os.Open(localPath)
	if err != nil {
		return file, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	return file, nil
}

func (s *Storage) GetImageFromS3(imagePath string) (string, error) {
	localImagePath := path.Join("s3", filepath.Base(imagePath))
	err := s.DownloadS3File(context.Background(), imagePath, localImagePath)
	if err != nil {
		log.Printf("Failed to download image: %v", err)
		return localImagePath, err
	}

	localImagePath = path.Join("/", localImagePath)

	return localImagePath, nil
}

func (s *Storage) Health() map[string]string {
	health := make(map[string]string)

	_, err := s.S3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
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
