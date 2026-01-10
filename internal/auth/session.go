package auth

import (
	"fmt"
	"maps"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

// SessionStore wraps sessions.CookieStore to allow custom methods.
type SessionStore struct {
	*sessions.CookieStore

	sessionName string
}

// InitializeSession initializes the session store with options.
func InitializeSession(key string) *SessionStore {
	if key == "" {
		key = os.Getenv("SESSION_NAME")
	}

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
		key,
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
