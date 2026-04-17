package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"gopkg.in/yaml.v2"
)

// Storage provides storage operations with support for S3 and local filesystem.
type Storage struct {
	UseS3      bool
	BucketName string
	BaseDir    string // Directory for local storage, defaults to "storage"
	PromptsDir string // Directory for prompt files, defaults to "prompts"
	S3Client   *s3.Client
}

// NewStorage initializes a new Storage instance.
func NewStorage(ctx context.Context) (*Storage, error) {
	// The USE_S3 environment variable determines whether to use S3 or local storage.
	useS3 := os.Getenv("USE_S3") == "true"
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	region := os.Getenv("AWS_REGION")

	var baseDir string

	// Find project root by looking for go.mod
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	baseDir = filepath.Join(projectRoot, "storage")
	promptsDir := filepath.Join(projectRoot, "prompts")

	// Verify storage directory exists
	_, err = os.Stat(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("storage directory not found at %s", baseDir)
		}

		return nil, fmt.Errorf("failed to check storage directory: %w", err)
	}

	if useS3 {
		if bucketName == "" || region == "" {
			return nil, errors.New("AWS_BUCKET_NAME or AWS_REGION is not set in the environment variables")
		}

		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			return nil, fmt.Errorf("unable to load SDK config, %w", err)
		}

		client := s3.NewFromConfig(cfg)

		return &Storage{
			UseS3:      true,
			BucketName: bucketName,
			BaseDir:    baseDir,
			PromptsDir: promptsDir,
			S3Client:   client,
		}, nil
	}

	return &Storage{
		UseS3:      false,
		BucketName: "",
		BaseDir:    baseDir,
		PromptsDir: promptsDir,
		S3Client:   nil,
	}, nil
}

// ListObjects lists the objects in the storage.
func (s *Storage) ListObjects(ctx context.Context, prefix string) ([]types.Object, error) {
	if s.UseS3 {
		return s.listS3Objects(ctx, prefix)
	}

	return s.listLocalObjects(prefix)
}

func (s *Storage) listS3Objects(ctx context.Context, prefix string) ([]types.Object, error) {
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

				return nil, noBucket
			}

			return nil, fmt.Errorf("listing S3 objects: %w", err)
		}

		objects = append(objects, output.Contents...)
	}

	// Default sort objects by date, from most current to least current
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(*objects[j].LastModified)
	})

	return objects, nil
}

// listLocalObjects lists objects previously pulled from S3 or stored manually.
// This doesn't recursively retrieve directories, only the files from the passed in directory.
// Subdirectories are ignored.
func (s *Storage) listLocalObjects(prefix string) ([]types.Object, error) {
	fullPath, err := LocalPath(s.BaseDir, prefix)
	if err != nil {
		return nil, fmt.Errorf("getting local path: %w", err)
	}

	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(fullPath, 0750)
		if err != nil {
			return nil, fmt.Errorf("creating local storage directory: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("checking local storage directory: %w", err)
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("reading local storage directory: %w", err)
	}

	objects := make([]types.Object, 0, len(entries))
	for _, entry := range entries {
		// Use IsDir() to filter out directories
		if entry.IsDir() {
			continue
		}

		// File info for size and mod time
		fileInfo, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("getting file info: %w", err)
		}

		keyPath, err := LocalPath(prefix, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("getting local path: %w", err)
		}

		objects = append(objects, types.Object{
			Key:          aws.String(keyPath),
			LastModified: aws.Time(fileInfo.ModTime()),
			Size:         aws.Int64(fileInfo.Size()),
		})
	}

	// Sort by date desc
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(*objects[j].LastModified)
	})

	return objects, nil
}

// DownloadS3File downloads a file from S3 to local storage.
func (s *Storage) DownloadS3File(ctx context.Context, objectKey string) error {
	if !s.UseS3 {
		// In local mode, no action needed as files are already local
		return nil
	}

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
			log.Printf("Failed to close S3 object body: %v", err)
		}
	}()

	fileName, err := LocalPath(s.BaseDir, objectKey)
	if err != nil {
		return fmt.Errorf("getting local path: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(fileName), 0750)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// #nosec G304 -- fileName is validated by LocalPath to prevent path traversal
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, result.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (s *Storage) UploadFileToS3(ctx context.Context, objectKey string) error {
	if !s.UseS3 {
		return errors.New("storage is configured to be local, not configured to use S3")
	}

	fileName, err := LocalPath(s.BaseDir, objectKey)
	if err != nil {
		return fmt.Errorf("getting local path: %w", err)
	}

	// #nosec G304 -- fileName is validated by LocalPath to prevent path traversal
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// GetPreparedFile retrieves a file and decodes it.
func (s *Storage) GetPreparedFile(ctx context.Context, key string, document any) error {
	file, err := s.GetFile(ctx, key)
	if err != nil {
		return err
	}
	defer file.Close()

	err = DecodeFile(file, document)
	if err != nil {
		return fmt.Errorf("failed to decode %s: %w", key, err)
	}

	return nil
}

// GetRawFile retrieves a YAML file and decodes it into the document struct.
// Use this when you need the metadata from a YAML file (e.g. for editing forms).
func (s *Storage) GetRawFile(ctx context.Context, key string, document any) error {
	file, err := s.GetFile(ctx, key)
	if err != nil {
		return err
	}
	defer file.Close()

	err = DecodeFile(file, document)
	if err != nil {
		return fmt.Errorf("failed to decode %s: %w", key, err)
	}

	return nil
}

// GetFile ensures the file is available locally at localPath.
func (s *Storage) GetFile(ctx context.Context, key string) (*os.File, error) {
	localPath, err := LocalPath(s.BaseDir, key)
	if err != nil {
		return nil, fmt.Errorf("getting local path: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(localPath), 0750)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if s.UseS3 {
		err := s.DownloadS3File(ctx, key)
		if err != nil {
			return nil, err
		}
	}

	// #nosec G304 -- localPath is validated by LocalPath to prevent path traversal
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// GetImage downloads an image and returns its local path (or URL path).
func (s *Storage) GetImage(ctx context.Context, imageName string) (string, error) {
	// imageName is expected to include the subdirectory
	localImagePath, err := LocalPath(s.BaseDir, imageName)
	if err != nil {
		return "", err
	}

	if s.UseS3 {
		err := s.DownloadS3File(ctx, imageName)
		if err != nil {
			log.Printf("Failed to download image: %v", err)

			return localImagePath, err
		}
	}

	// Return the URL path for accessing the image
	webPath := "/storage/" + imageName

	return webPath, nil
}

// Health checks the storage connection status.
func (s *Storage) Health() map[string]string {
	health := make(map[string]string)

	if s.UseS3 {
		_, err := s.S3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
		if err != nil {
			health["status"] = "down"
			health["message"] = fmt.Sprintf("Failed to list buckets: %v", err)
			log.Printf("S3 connection down: %v", err) // Changed from Fatalf to allow app to survive

			return health
		}

		health["status"] = "up"
		health["message"] = "S3 storage is up and running."
	} else {
		_, err := os.Stat(s.BaseDir)
		if os.IsNotExist(err) {
			health["status"] = "down"
			health["message"] = "Local storage directory missing"
		} else {
			health["status"] = "up"
			health["message"] = "Local storage is ready."
		}
	}

	return health
}

// findProjectRoot walks up the directory tree to find the project root based on go.mod.
func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	dir := cwd
	for {
		// Check if go.mod exists in current directory
		goModPath := filepath.Join(dir, "go.mod")

		_, err := os.Stat(goModPath)
		if err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			return "", errors.New("could not find project root (go.mod)")
		}

		dir = parent
	}
}

// LocalPath checks validity and security of filename, returning the full local path.
func LocalPath(path, filename string) (string, error) {
	fp := filepath.Join(path, filename)
	cleaned := filepath.Clean(fp)

	// IsLocal ensures filename is within the subtree, not absolute, and not empty.
	if !filepath.IsLocal(filename) {
		return "", fmt.Errorf("invalid filepath: %s", filename)
	}

	return cleaned, nil
}

// --- Helpers ---

// DecodeFile decodes a YAML file into the provided output structure.
func DecodeFile(file io.Reader, out any) error {
	decoder := yaml.NewDecoder(file)

	err := decoder.Decode(out)
	if err != nil {
		log.Printf("Failed to decode file: %v", err)

		return fmt.Errorf("decode error: %w", err)
	}

	return nil
}

// GetDocumentBodyRaw reads the Markdown body file paired with yamlKey and returns raw markdown.
// The body file is expected at the same path as yamlKey but with a .md extension.
func (s *Storage) GetDocumentBodyRaw(ctx context.Context, yamlKey string) (string, error) {
	mdKey := strings.TrimSuffix(yamlKey, ".yaml") + ".md"

	file, err := s.GetFile(ctx, mdKey)
	if err != nil {
		return "", fmt.Errorf("failed to get body file %s: %w", mdKey, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	return string(content), nil
}

// GetDocumentBody reads the Markdown body file paired with yamlKey, converts it to HTML, and returns it.
// The body file is expected at the same path as yamlKey but with a .md extension.
func (s *Storage) GetDocumentBody(ctx context.Context, yamlKey string) (string, error) {
	mdKey := strings.TrimSuffix(yamlKey, ".yaml") + ".md"

	file, err := s.GetFile(ctx, mdKey)
	if err != nil {
		return "", fmt.Errorf("failed to get body file %s: %w", mdKey, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	html, err := MarkdownToHTML(content)
	if err != nil {
		return "", err
	}

	return html, nil
}

// FormatFileSize formats a byte count as a human-readable string.
func FormatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}

	return fmt.Sprintf("%.1f KB", float64(size)/1024)
}

// WriteMarkdownDocument writes a document as two separate files:
// - yamlPath: YAML metadata file containing title, subtitle, tags, author, etc.
// - mdPath: Markdown body file containing the document content.
// The "body" key in formData is written to mdPath; all other keys are written as YAML to yamlPath.
func WriteMarkdownDocument(yamlPath, mdPath string, formData map[string]any) error {
	err := os.MkdirAll(filepath.Dir(yamlPath), 0750)
	if err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", yamlPath, err)
	}

	var body string

	if bodyVal, exists := formData["body"]; exists {
		var ok bool

		body, ok = bodyVal.(string)
		if !ok {
			return fmt.Errorf("WriteMarkdownDocument: body must be a string, got %T", bodyVal)
		}
	}

	metaData := make(map[string]any, len(formData))
	for k, v := range formData {
		if k != "body" {
			metaData[k] = v
		}
	}

	fm, err := yaml.Marshal(metaData)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// #nosec G304 -- yamlPath comes from internal code paths, validated by callers using LocalPath
	yf, err := os.Create(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to create yaml file: %w", err)
	}
	defer yf.Close()

	_, err = yf.Write(fm)
	if err != nil {
		return fmt.Errorf("failed to write yaml file: %w", err)
	}

	// #nosec G304 -- mdPath comes from internal code paths, validated by callers using LocalPath
	mf, err := os.Create(mdPath)
	if err != nil {
		return fmt.Errorf("failed to create markdown file: %w", err)
	}
	defer mf.Close()

	title, _ := formData["title"].(string)
	subtitle, _ := formData["subtitle"].(string)

	_, err = fmt.Fprintf(mf, "# %s\n## %s\n\n%s", title, subtitle, body)
	if err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

// GetPromptContent returns the system prompt content for the given document type.
// docType should be one of: "articles", "projects", "reading-list", "letters".
func (s *Storage) GetPromptContent(ctx context.Context, docType string) (string, error) {
	// Validate docType against allowlist to prevent directory traversal
	validDocTypes := map[string]string{
		"articles":     "articles.txt",
		"projects":     "projects.txt",
		"reading-list": "reading-list.txt",
		"letters":      "letters.txt",
	}

	filename, exists := validDocTypes[docType]
	if !exists {
		return "", fmt.Errorf("unsupported document type: %q", docType)
	}

	if s.UseS3 {
		// Read from S3
		key := filename

		result, err := s.S3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.BucketName),
			Key:    aws.String(key),
		})
		if err != nil {
			return "", fmt.Errorf("failed to read prompt from S3: %w", err)
		}

		defer func() {
			cerr := result.Body.Close()
			if cerr != nil {
				log.Printf("failed to close S3 response body: %v", cerr)
			}
		}()

		content, err := io.ReadAll(result.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read S3 object: %w", err)
		}

		return string(content), nil
	}

	// Read from local filesystem
	content, err := fs.ReadFile(os.DirFS(s.PromptsDir), filename)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %q: %w", filename, err)
	}

	return string(content), nil
}
