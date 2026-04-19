package web

import (
	"log"
	"net/http"

	apperrors "timterests/internal/errors"

	"timterests/internal/auth"
)

func LoginHandler(w http.ResponseWriter, r *http.Request, a *auth.Auth) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		authenticated, err := a.Authenticate(r.Context(), w, r, email, password)
		if authenticated && err == nil {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)

			return
		}

		log.Printf("login: authentication failed: %v", err)

		if err != nil && !apperrors.Is(err, auth.ErrInvalidCredentials) {
			HandleError(w, r, apperrors.InternalServerError(err), "LoginHandler", "authenticate")

			return
		}

		component := LoginPage("Incorrect email or password.")
		if r.Header.Get("Hx-Request") == "true" {
			component = LoginContainer("Incorrect email or password.")
		}

		renderErr := renderHTML(w, r, http.StatusUnauthorized, component)
		if renderErr != nil {
			HandleError(w, r, apperrors.RenderFailed(renderErr), "LoginHandler", "renderError")
		}

		return
	}

	component := LoginPage("")
	if r.Header.Get("Hx-Request") == "true" {
		component = LoginContainer("")
	}

	err := renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "LoginHandler", "render")
	}
}
