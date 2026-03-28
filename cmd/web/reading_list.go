package web

import (
	"fmt"
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
		message := "Failed to fetch reading list"
		http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusInternalServerError)

		return
	}

	for i := range books {
		books[i].Body = storage.RemoveHTMLTags(books[i].Body)
		v := reflect.ValueOf(books[i])
		tags = storage.GetTags(v, tags)
	}

	if IsHTMXRequest(r) {
		SetPartialResponseHeaders(w)
		component = ReadingListList(books, design)
	} else {
		component = ReadingListPage(books, tags, design)
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

	books, err := service.ListBooks(r.Context(), s, "all")
	if err != nil {
		http.Error(w, "Failed to fetch books", http.StatusInternalServerError)

		return
	}

	for _, book := range books {
		if book.ID == bookID {
			authenticated := a.IsAuthenticated(r)

			if IsHTMXRequest(r) {
				SetPartialResponseHeaders(w)
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
