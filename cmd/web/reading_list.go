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

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		log.Printf("ReadingListPageHandler: failed to render: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetReadingListBook retrieves and renders a specific book by ID.
func GetReadingListBook(w http.ResponseWriter, r *http.Request, s storage.Storage, bookID string, a *auth.Auth) {
	books, err := service.ListBooks(r.Context(), s, "all")
	if err != nil {
		http.Error(w, "Failed to fetch books", http.StatusInternalServerError)

		return
	}

	for _, book := range books {
		if book.ID == bookID {
			var component templ.Component

			authenticated := a.IsAuthenticated(r)

			if r.Header.Get("Hx-Request") == "true" {
				component = BookDisplay(book, authenticated)
			} else {
				component = BookPage(book, authenticated)
			}

			err = renderHTML(w, r, http.StatusOK, component)
			if err != nil {
				log.Printf("GetReadingListBook: failed to render: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			return
		}
	}

	http.Error(w, "Not Found", http.StatusNotFound)
}
