package auth

import (
	"fmt"
	"maps"
	"net/http"

	"github.com/gorilla/sessions"
)

// SessionStore wraps sessions.CookieStore to allow custom methods.
type SessionStore struct {
	*sessions.CookieStore

	sessionName string
}

// MinSessionKeyLength is the shortest accepted signing key. Anything weaker and
// a forged cookie becomes a realistic route straight past authentication.
const MinSessionKeyLength = 32

// InitializeSession initializes the session store with options.
//
// name identifies the cookie and is public. key signs it and is a secret — they
// are deliberately separate, because a value that appears in a cookie name can
// never be trusted to protect its contents.
func InitializeSession(name, key string) *SessionStore {
	store := sessions.NewCookieStore([]byte(key))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	sess := &SessionStore{
		store,
		name,
	}

	return sess
}

// SetSessionValue sets a map of key-value pairs to a global session.
func (store *SessionStore) SetSessionValue(w http.ResponseWriter, r *http.Request, values map[any]any) error {
	session, err := store.Get(r, store.sessionName)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	maps.Copy(session.Values, values)

	err = session.Save(r, w)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// ClearSession drops every value and expires the cookie.
func (store *SessionStore) ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, store.sessionName)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	clear(session.Values)

	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		return fmt.Errorf("failed to clear session: %w", err)
	}

	return nil
}

// GetSessionValue retrieves a value from the session using the provided key.
func (store *SessionStore) GetSessionValue(r *http.Request, key any) string {
	session, err := store.Get(r, store.sessionName)
	if err != nil {
		return ""
	}

	value := session.Values[key]
	if value == nil {
		return ""
	}

	str, ok := value.(string)
	if !ok {
		return ""
	}

	return str
}
