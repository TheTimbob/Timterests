package web

import (
	"net/http"
	"timterests/internal/auth"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Perform authentication
		authenticated, err := auth.Authenticate(w, r, email, password)
		if authenticated && err == nil {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		} else {
			http.Error(w, "Authentication: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Render the login page for the initial load
	component := LoginPage()
	if r.Header.Get("HX-Request") == "true" {
		// Render the login container for main inner html replacement
		component = LoginContainer()
	}

	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
