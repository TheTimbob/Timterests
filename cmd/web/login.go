package web

import (
	"errors"
	"log"
	"net/http"
	"timterests/internal/auth"
)

// LoginHandler handles user authentication and login requests.
func LoginHandler(w http.ResponseWriter, r *http.Request, a *auth.Auth) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		authenticated, err := a.Authenticate(r.Context(), w, r, email, password)
		if authenticated && err == nil {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)

			return
		}

		log.Printf("login: authentication failed for %q: %v", email, err)

		// Distinguish invalid credentials (401) from unexpected server errors (500).
		if err != nil && !errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}

		// Re-render the login page with an inline error rather than navigating away.
		component := LoginPage("Incorrect email or password.")
		if r.Header.Get("Hx-Request") == "true" {
			component = LoginContainer("Incorrect email or password.")
		}

		renderErr := renderHTML(w, r, http.StatusUnauthorized, component)
		if renderErr != nil {
			log.Printf("login: failed to render error page: %v", renderErr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		return
	}

	// Render the login page for the initial load.
	component := LoginPage("")
	if r.Header.Get("Hx-Request") == "true" {
		component = LoginContainer("")
	}

	err := renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		log.Printf("login: failed to render page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
