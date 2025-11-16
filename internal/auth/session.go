package auth

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

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

// Sets a map of key-value pairs to a global session using the provided key.
func SetSessionValue(w http.ResponseWriter, r *http.Request, values map[interface{}]interface{}) error {
	session, err := store.Get(r, os.Getenv("SESSION_NAME"))
	if err != nil {
		return err
	}

	for key, value := range values {
		session.Values[key] = value
	}

	err = session.Save(r, w)
	if err != nil {
		return err
	}
	return nil
}

// Gets a value from the session using the provided key.
func GetSessionValue(r *http.Request, key interface{}) string {
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
