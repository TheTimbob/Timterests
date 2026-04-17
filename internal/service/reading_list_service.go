package service

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"timterests/internal/model"
	"timterests/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// ListBooks retrieves all books from the reading list in storage,
// optionally filtering by tag. Pass tag="" or tag="all" to retrieve all books.
func ListBooks(ctx context.Context, s storage.Storage, tag string) ([]model.ReadingList, error) {
	var readingList []model.ReadingList

	prefix := "reading-list/"

	files, err := s.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	mdKeys := make(map[string]bool, len(files))
	for _, obj := range files {
		key := aws.ToString(obj.Key)
		if strings.HasSuffix(key, ".md") {
			mdKeys[key] = true
		}
	}

	docIdx := 0

	for _, obj := range files {
		key := aws.ToString(obj.Key)

		if key == prefix || !strings.HasSuffix(key, ".yaml") {
			continue
		}

		if !mdKeys[strings.TrimSuffix(key, ".yaml")+".md"] {
			log.Printf("ListBooks: skipping %s — no paired .md body file", key)

			continue
		}

		book, err := GetBook(ctx, s, key, docIdx)
		if err != nil {
			return nil, err
		}

		docIdx++

		if tag == "all" || tag == "" || slices.Contains(book.Tags, tag) {
			readingList = append(readingList, *book)
		}
	}

	return readingList, nil
}

// GetBook retrieves a single book by its storage key and numeric ID,
// including downloading and resolving its associated cover image.
func GetBook(ctx context.Context, s storage.Storage, key string, id int) (*model.ReadingList, error) {
	book, err := getDoc[model.ReadingList](ctx, s, key, id)
	if err != nil {
		return nil, err
	}

	imagePath, err := s.GetImage(ctx, book.Image)
	if err != nil {
		log.Printf("Failed to download image: %v", err)

		return nil, fmt.Errorf("failed to resolve image %q: %w", book.Image, err)
	}

	book.Image = imagePath

	return book, nil
}
