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
	"timterests/internal/models"
	"timterests/internal/storage"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func ReadingListPageHandler(w http.ResponseWriter, r *http.Request, storageInstance models.Storage, currentTag, design string) {
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
		component = ReadingList(readingList, design)
	} else {
		component = ReadingListPage(readingList, tags, currentTag, design)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in ReadingListHandler: %e", err)
	}
}

func GetReadingListBook(w http.ResponseWriter, r *http.Request, storageInstance models.Storage, bookID string) {

	readingList, err := ListBooks(storageInstance, "all")
	if err != nil {
		http.Error(w, "Failed to fetch books", http.StatusInternalServerError)
		return
	}

	for _, book := range readingList {
		if book.ID == bookID {
			render := true
			component := BookPage(book, render)
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Fatalf("Error rendering in GetReadingListBook: %e", err)
			}
		}
	}
}

func ListBooks(storageInstance models.Storage, tag string) ([]models.ReadingList, error) {
	var readingList []models.ReadingList

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

func GetBook(key string, id int, storageInstance models.Storage) (*models.ReadingList, error) {
	var book models.ReadingList
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
