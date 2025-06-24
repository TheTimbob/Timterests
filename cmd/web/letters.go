package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
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

func (l Letter) GetID() string {
	return l.ID
}

func (l Letter) GetBody() string {
	return l.Body
}

func (l Letter) GetTitle() string {
	return l.Title
}

func (l Letter) GetSubtitle() string {
	return l.Subtitle
}

func (l Letter) GetTags() []string {
	return l.Tags
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

	if currentTag != "" || design != "" {
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
	var component templ.Component = nil

	// Check if user is authenticated
	if !IsAuthenticated(r) {
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
			if r.Header.Get("HX-Request") == "true" {
				component = LetterDisplay(letter)
			} else {
				component = LetterPage(letter)
			}
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Printf("Error rendering in GetLettereByIDHandler: %e", err)
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
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	// Retrieve file content
	file, err := storage.GetFile(key, localFilePath, storageInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", key, err)
	}

	if err := storage.DecodeFile(file, &letter); err != nil {
		return nil, fmt.Errorf("failed to decode file %s: %w", key, err)
	}

	body, err := storage.BodyToHTML(letter.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the body text into HTML: %w", err)
	}

	letter.Body = body
	letter.ID = strconv.Itoa(id)
	return &letter, nil
}
