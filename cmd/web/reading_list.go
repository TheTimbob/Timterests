package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"reflect"
	"slices"
	"strconv"
	"timterests/internal/storage"
	"timterests/internal/types"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type ReadingList struct {
	types.Document `yaml:",inline"`
	Image          string `yaml:"image-path"`
	Author         string `yaml:"author"`
	Published      string `yaml:"published"`
	ISBN           string `yaml:"isbn"`
	Website        string `yaml:"website"`
	Status         string `yaml:"status"`
}

func ReadingListPageHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, currentTag, design string) {
	var component templ.Component
	var tags []string

	readingList, err := ListBooks(storageInstance, currentTag)
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

	if currentTag != "" || design != "" {
		component = ReadingListList(readingList, design)
	} else {
		component = ReadingListPage(readingList, tags, design)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in ReadingListHandler: %e", err)
	}
}

func GetReadingListBook(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, bookID string) {

	readingList, err := ListBooks(storageInstance, "all")
	if err != nil {
		http.Error(w, "Failed to fetch books", http.StatusInternalServerError)
		return
	}

	for _, book := range readingList {
		if book.ID == bookID {
			component := BookPage(book)
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Fatalf("Error rendering in GetReadingListBook: %e", err)
			}
		}
	}
}

func ListBooks(storageInstance storage.Storage, tag string) ([]ReadingList, error) {
	var readingList []ReadingList

	// Get all readingList from the storage
	prefix := "reading-list/"
	files, err := storage.ListObjects(context.Background(), storageInstance, prefix)
	if err != nil {
		return nil, err
	}

	for id, obj := range files {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		book, err := GetBook(key, id, storageInstance)
		if err != nil {
			return nil, err
		}

		if slices.Contains(book.Tags, tag) || tag == "all" || tag == "" {
			readingList = append(readingList, *book)
		}
	}

	return readingList, nil
}

func GetBook(key string, id int, storageInstance storage.Storage) (*ReadingList, error) {
	var book ReadingList
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	file, err := storage.GetFile(key, localFilePath, storageInstance)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
		return nil, err
	}

	if err := storage.DecodeFile(file, &book); err != nil {
		log.Fatalf("Failed to decode file: %v", err)
		return nil, err
	}

	body, err := storage.BodyToHTML(book.Body)
	if err != nil {
		log.Fatalf("Failed to parse the body into HTML: %v", err)
		return nil, err
	}

	localImagePath, err := storage.GetImageFromS3(storageInstance, book.Image)
	if err != nil {
		log.Fatalf("Failed to download image: %v", err)
		return nil, err
	}

	book.Image = localImagePath
	book.Body = body
	book.ID = strconv.Itoa(id)
	return &book, nil
}
