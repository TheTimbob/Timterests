// Package auth provides authentication and session management functionality.
package auth

import (
	"net/http"
)

// Auth provides authentication and session management functionality.
type Auth struct {
	store *SessionStore
}

// NewAuth creates a new Auth instance. The name identifies the session cookie;
// the key signs it and must be a secret of at least MinSessionKeyLength.
func NewAuth(sessionName, sessionKey string) *Auth {
	return &Auth{
		store: InitializeSession(sessionName, sessionKey),
	}
}

// IsAuthenticated checks if the user is authenticated based on session values.
func (a *Auth) IsAuthenticated(r *http.Request) bool {
	// Check if the user is authenticated
	session := a.store.GetSessionValue(r, "email")

	return session != ""
}

// ClearSession signs the user out by dropping the session.
func (a *Auth) ClearSession(w http.ResponseWriter, r *http.Request) error {
	return a.store.ClearSession(w, r)
}

// SetSessionValue sets session values. This is primarily used for testing.
func (a *Auth) SetSessionValue(w http.ResponseWriter, r *http.Request, values map[any]any) error {
	return a.store.SetSessionValue(w, r, values)
}
