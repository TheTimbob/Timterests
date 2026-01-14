package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"timterests/cmd/web/components"
	"timterests/internal/auth"
	"timterests/internal/model"
	"timterests/internal/storage"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// Letter represents a personal letter with date and occasion.
type Letter struct {
	model.Document `yaml:",inline"`

	Date     string `yaml:"date"`
	Occasion string `yaml:"occasion"`
}

// LettersPageHandler handles requests to the letters page,
// ensuring authentication and rendering the appropriate content.
func LettersPageHandler(
	w http.ResponseWriter,
	r *http.Request,
	s storage.Storage,
	currentTag, design string,
	a *auth.Auth) {
	var (
		component templ.Component
		tags      []string
	)

	// Check if user is authenticated

	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	letters, err := ListLetters(r.Context(), s, currentTag)
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

	if r.Header.Get("Hx-Request") == "true" {
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

// GetLetterHandler retrieves a specific letter by its ID and renders it.
func GetLetterHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, letterID string, a *auth.Auth) {
	// Check if user is authenticated
	authenticated := a.IsAuthenticated(r)
	if !authenticated {
		log.Printf("User not authenticated, redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	letters, err := ListLetters(r.Context(), s, "all")
	if err != nil {
		http.Error(w, "Failed to fetch letters", http.StatusInternalServerError)

		return
	}

	for _, letter := range letters {
		if letter.ID == letterID {
			var component templ.Component
			if r.Header.Get("Hx-Request") == "true" {
				component = LetterDisplay(letter, authenticated)
			} else {
				component = LetterPage(letter, authenticated)
			}

			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Printf("Error rendering in GetLetterByIDHandler: %e", err)
			}
		}
	}
}

// ListLetters retrieves all letters from storage, and returns a slice of Letter structs.
func ListLetters(ctx context.Context, s storage.Storage, tag string) ([]Letter, error) {
	// Get all letters from the storage
	prefix := "letters/"

	letterFiles, err := s.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	letters := make([]Letter, 0, len(letterFiles))

	for id, obj := range letterFiles {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		letter, err := GetLetter(ctx, key, id, s)
		if err != nil {
			return nil, err
		}

		if slices.Contains(letter.Tags, tag) || tag == "all" || tag == "" {
			letters = append(letters, *letter)
		}
	}

	sort.Slice(letters, func(i, j int) bool {
		return letters[i].Date > letters[j].Date
	})

	return letters, nil
}

// GetLetter retrieves a single letter by its S3 key and ID.
func GetLetter(ctx context.Context, key string, id int, s storage.Storage) (*Letter, error) {
	var letter Letter

	letter.ID = strconv.Itoa(id)
	letter.S3Key = key

	err := s.GetPreparedFile(ctx, key, &letter)
	if err != nil {
		return nil, fmt.Errorf("failed to get prepared file: %w", err)
	}

	return &letter, nil
}

// ToCard converts a Letter to a Card component for display in lists.
func (l Letter) ToCard(i int) components.Card {
	return components.Card{
		Title:     l.Title,
		Subtitle:  l.Subtitle,
		Date:      l.Date,
		Body:      l.Body,
		ImagePath: "",
		Get:       "/letter?id=" + l.ID,
		Tags:      l.Tags,
		Index:     i,
	}
}
