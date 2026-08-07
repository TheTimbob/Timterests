package web

import (
	"net/http"

	apperrors "timterests/internal/errors"
)

// LoginHandler renders the sign-in page. Authentication itself happens through
// Cognito; see OIDCLoginHandler.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	component := LoginPage("")

	if IsHTMXRequest(r) {
		SetPartialResponseHeaders(w)

		component = LoginContainer("")
	}

	err := renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "LoginHandler", "render")
	}
}
