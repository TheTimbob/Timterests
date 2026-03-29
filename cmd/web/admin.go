package web

import (
	"log"
	"net/http"
	"timterests/internal/auth"

	"github.com/a-h/templ"
)

// AdminPageHandler handles requests to the admin page for authenticated users.
func AdminPageHandler(w http.ResponseWriter, r *http.Request, a *auth.Auth) {
	var component templ.Component

	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	component = AdminPage()

	err := renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		log.Printf("AdminPageHandler: failed to render: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
