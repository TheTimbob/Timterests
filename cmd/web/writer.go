package web

import (
	"log"
	"net/http"

	"github.com/a-h/templ"
)

func WriterPageHandler(w http.ResponseWriter, r *http.Request, docType string) {
	var component templ.Component

	// TODO - Uncomment this before release
	// if !IsAuthenticated(r) {
	// 	http.Redirect(w, r, "/login", http.StatusSeeOther)
	// 	return
	// }

	if r.Header.Get("HX-Request") == "true" {
		component = FormContentByType(docType)
	} else {
		component = WriterPage()
	}

	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in WriterPage: %v", err)
	}
}
