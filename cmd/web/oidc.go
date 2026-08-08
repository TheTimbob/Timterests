package web

import (
	"errors"
	"log"
	"net/http"

	"timterests/internal/auth"
	apperrors "timterests/internal/errors"
)

// OIDCEnabled reports whether Cognito sign-in is configured, so the login page
// can offer it without needing the handler plumbed into the template.
func OIDCEnabled() bool {
	return auth.OIDCConfigFromEnv().Configured()
}

// OIDCLoginHandler starts the Cognito handshake.
func OIDCLoginHandler(w http.ResponseWriter, r *http.Request, o *auth.OIDC) {
	if !o.Configured() {
		http.NotFound(w, r)

		return
	}

	redirectURL, err := o.AuthCodeURL(w, r)
	if err != nil {
		log.Printf("oidc: failed to build auth URL: %v", err)
		HandleError(w, r, apperrors.InternalServerError(err), "OIDCLoginHandler", "authCodeURL")

		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// OIDCCallbackHandler completes the handshake and signs the user in.
func OIDCCallbackHandler(w http.ResponseWriter, r *http.Request, o *auth.OIDC) {
	if !o.Configured() {
		http.NotFound(w, r)

		return
	}

	err := o.Complete(w, r)
	if err == nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)

		return
	}

	// A rejected account is an expected outcome, not a fault. Anything else is
	// logged as a failure but still shown as a generic sign-in error, so the page
	// never leaks why the handshake broke.
	if errors.Is(err, auth.ErrNotAuthorized) {
		renderLoginError(w, r, "That account is not authorized.", http.StatusForbidden)

		return
	}

	log.Printf("oidc: sign-in failed: %v", err)
	renderLoginError(w, r, "Sign-in failed. Please try again.", http.StatusUnauthorized)
}

// LogoutHandler clears the local session, then hands off to Cognito so the
// upstream session is dropped too. Without that second step the next sign-in
// would silently reuse the existing Cognito session.
func LogoutHandler(w http.ResponseWriter, r *http.Request, a *auth.Auth, o *auth.OIDC) {
	err := a.ClearSession(w, r)
	if err != nil {
		log.Printf("logout: failed to clear session: %v", err)
	}

	if o.Configured() {
		http.Redirect(w, r, o.LogoutURL(), http.StatusSeeOther)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func renderLoginError(w http.ResponseWriter, r *http.Request, message string, status int) {
	component := LoginPage(message)

	if IsHTMXRequest(r) {
		SetPartialResponseHeaders(w)

		component = LoginContainer(message)
	}

	err := renderHTML(w, r, status, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "renderLoginError", "render")
	}
}
