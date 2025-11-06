package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"timterests/internal/storage"
	"timterests/internal/types"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type Letter struct {
	types.Document `yaml:",inline"`
	Date           string `yaml:"date"`
	Occasion       string `yaml:"occasion"`
}

func LettersPageHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, currentTag, design string) {
	var component templ.Component
	var tags []string

	// Check if user is authenticated
	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	letters, err := ListLetters(storageInstance)
	if err != nil {
		message := "Failed to fetch letters"
		http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusInternalServerError)
		return
	}

	for i := range letters {
		letters[i].Body = storage.RemoveHTMLTags(letters[i].Body)
		v := reflect.ValueOf(letters[i])
		tags = storage.GetTags(v, tags)
	}

	if r.Header.Get("HX-Request") == "true" {
		component = LettersList(letters, design)
	} else {
		component = LettersListPage(letters, tags, design)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in LettersPosts: %e", err)
	}
}

func GetLetterHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, letterID string) {
	// Check if user is authenticated
	isAuthenticated := IsAuthenticated(r)
	if !isAuthenticated {
		log.Printf("User not authenticated, redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	letters, err := ListLetters(storageInstance)
	if err != nil {
		http.Error(w, "Failed to fetch letters", http.StatusInternalServerError)
		return
	}

	for _, letter := range letters {
		if letter.ID == letterID {
			var component templ.Component
			if r.Header.Get("HX-Request") == "true" {
				component = LetterDisplay(letter, isAuthenticated)
			} else {
				component = LetterPage(letter, isAuthenticated)
			}
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Printf("Error rendering in GetLetterByIDHandler: %e", err)
			}
		}
	}

}

func ListLetters(storageInstance storage.Storage) ([]Letter, error) {
	var letters []Letter

	// Get all letters from the storage
	prefix := "letters/"
	letterFiles, err := storage.ListObjects(context.Background(), storageInstance, prefix)
	if err != nil {
		return nil, err
	}

	for id, obj := range letterFiles {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		letter, err := GetLetter(key, id, storageInstance)
		if err != nil {
			return nil, err
		}

		letters = append(letters, *letter)
	}

	sort.Slice(letters, func(i, j int) bool {
		return letters[i].Date > letters[j].Date
	})

	return letters, nil
}

func GetLetter(key string, id int, storageInstance storage.Storage) (*Letter, error) {
	var letter Letter
	letter.ID = strconv.Itoa(id)
	letter.S3Key = key
	err := storage.GetPreparedFile(key, &letter, storageInstance)
	if err != nil {
		return nil, err
	}

	return &letter, nil
}
