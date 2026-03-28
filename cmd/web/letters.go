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

	letters, err := service.ListLetters(r.Context(), s, currentTag)
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

	if IsHTMXRequest(r) {
		SetPartialResponseHeaders(w)
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

	letters, err := service.ListLetters(r.Context(), s, "all")
	if err != nil {
		http.Error(w, "Failed to fetch letters", http.StatusInternalServerError)

		return
	}

	for _, letter := range letters {
		if letter.ID == letterID {
			var component templ.Component
			if IsHTMXRequest(r) {
				SetPartialResponseHeaders(w)
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
