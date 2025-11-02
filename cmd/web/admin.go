package web

import (
	"log"
	"net/http"

	"github.com/a-h/templ"
)

func AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	var component templ.Component

	if !IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	component = AdminPage()

	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in WriterPage: %v", err)
	}
}
