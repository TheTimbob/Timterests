package auth

import (
	"fmt"
	"maps"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

// session store needs to be global and initialized once at package level
//
//nolint:gochecknoglobals
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_NAME")))

// Initialize the session store with options.
func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

// SetSessionValue sets a map of key-value pairs to a global session.
func SetSessionValue(w http.ResponseWriter, r *http.Request, values map[any]any) error {
	session, err := store.Get(r, os.Getenv("SESSION_NAME"))
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
func GetSessionValue(r *http.Request, key any) string {
	session, err := store.Get(r, os.Getenv("SESSION_NAME"))
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
