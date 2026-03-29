package web

import (
	"log"
	"net/http"
	"reflect"
	"timterests/internal/auth"
	"timterests/internal/service"
	"timterests/internal/storage"

	"github.com/a-h/templ"
)

// ReadingListPageHandler handles requests to the reading list page and renders book collections.
func ReadingListPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, currentTag, design string) {
	var (
		component templ.Component
		tags      []string
	)

	books, err := service.ListBooks(r.Context(), s, currentTag)
	if err != nil {
		log.Printf("ReadingListPageHandler: failed to fetch reading list: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	for i := range books {
		books[i].Body = storage.RemoveHTMLTags(books[i].Body)
		v := reflect.ValueOf(books[i])
		tags = storage.GetTags(v, tags)
	}

	if r.Header.Get("Hx-Request") == "true" {
		component = ReadingListList(books, design)
	} else {
		component = ReadingListPage(books, tags, design)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		log.Printf("ReadingListPageHandler: failed to render: %v", err)
	}
}

// GetReadingListBook retrieves and renders a specific book by ID.
func GetReadingListBook(w http.ResponseWriter, r *http.Request, s storage.Storage, bookID string, a *auth.Auth) {
	var component templ.Component

	books, err := service.ListBooks(r.Context(), s, "all")
	if err != nil {
		http.Error(w, "Failed to fetch books", http.StatusInternalServerError)

		return
	}

	found := false

	for _, book := range books {
		if book.ID == bookID {
			found = true
			authenticated := a.IsAuthenticated(r)

			if r.Header.Get("Hx-Request") == "true" {
				component = BookDisplay(book, authenticated)
			} else {
				component = BookPage(book, authenticated)
			}

			err = component.Render(r.Context(), w)
			if err != nil {
				log.Printf("GetReadingListBook: failed to render: %v", err)
			}
		}
	}

	if !found {
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}
