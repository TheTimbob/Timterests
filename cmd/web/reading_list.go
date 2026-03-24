package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"timterests/cmd/web/components"
	"timterests/internal/auth"
	"timterests/internal/model"
	"timterests/internal/storage"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// ReadingList represents a book that appears in the reading list.
type ReadingList struct {
	model.Document `yaml:",inline"`

	Image     string `yaml:"imagePath"`
	Author    string `yaml:"author"`
	Published string `yaml:"published"`
	ISBN      string `yaml:"isbn"`
	Website   string `yaml:"website"`
	Status    string `yaml:"status"`
}

// ReadingListPageHandler handles requests to the reading list page and renders book collections.
func ReadingListPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, currentTag, design string) {
	var (
		component templ.Component
		tags      []string
	)

	readingList, err := ListBooks(r.Context(), s, currentTag)
	if err != nil {
		message := "Failed to fetch reading list"
		http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusInternalServerError)

		return
	}

	for i := range readingList {
		readingList[i].Body = storage.RemoveHTMLTags(readingList[i].Body)
		v := reflect.ValueOf(readingList[i])
		tags = storage.GetTags(v, tags)
	}

	if r.Header.Get("Hx-Request") == "true" {
		component = ReadingListList(readingList, design)
	} else {
		component = ReadingListPage(readingList, tags, design)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in ReadingListHandler: %e", err)
	}
}

// GetReadingListBook retrieves and renders a specific book by ID.
func GetReadingListBook(w http.ResponseWriter, r *http.Request, s storage.Storage, bookID string, a *auth.Auth) {
	var component templ.Component

	readingList, err := ListBooks(r.Context(), s, "all")
	if err != nil {
		http.Error(w, "Failed to fetch books", http.StatusInternalServerError)

		return
	}

	for _, book := range readingList {
		if book.ID == bookID {
			authenticated := a.IsAuthenticated(r)
			if r.Header.Get("Hx-Request") == "true" {
				component = BookDisplay(book, authenticated)
			} else {
				component = BookPage(book, authenticated)
			}

			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Printf("Error rendering in GetReadingListBook: %e", err)
			}
		}
	}
}

// ListBooks retrieves all books from storage, optionally filtered by tag.
func ListBooks(ctx context.Context, s storage.Storage, tag string) ([]ReadingList, error) {
	var readingList []ReadingList

	// Get all readingList from the storage
	prefix := "reading-list/"

	files, err := s.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	for id, obj := range files {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		book, err := GetBook(ctx, key, id, s)
		if err != nil {
			return nil, err
		}

		if slices.Contains(book.Tags, tag) || tag == "all" || tag == "" {
			readingList = append(readingList, *book)
		}
	}

	return readingList, nil
}

// GetBook retrieves a book by its S3 key and ID from storage.
func GetBook(ctx context.Context, key string, id int, s storage.Storage) (*ReadingList, error) {
	var book ReadingList

	book.SetMeta(strconv.Itoa(id), key)

	err := s.GetPreparedFile(ctx, key, &book)
	if err != nil {
		return nil, fmt.Errorf("failed to get prepared file: %w", err)
	}

	localImagePath, err := s.GetImage(ctx, book.Image)
	if err != nil {
		log.Printf("Failed to download image: %v", err)

		return nil, fmt.Errorf("failed to get image from S3: %w", err)
	}

	book.Image = localImagePath

	return &book, nil
}

// ToCard converts a ReadingList item to a Card component for display.
func (r ReadingList) ToCard(i int) components.Card {
	return components.Card{
		Title:     r.Title,
		Subtitle:  r.Subtitle,
		Date:      "",
		Body:      r.Body,
		ImagePath: r.Image,
		Get:       "/book?id=" + r.ID,
		Tags:      r.Tags,
		Index:     i,
	}
}
