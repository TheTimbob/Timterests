package web

import (
	"fmt"
	"net/http"

	"timterests/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Perform authentication
		isAuthenticated, err := Authenticate(w, r, email, password)
		if isAuthenticated && err == nil {
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

func Authenticate(w http.ResponseWriter, r *http.Request, email, password string) (bool, error) {

	db, err := storage.GetDB()
	if err != nil {
		return false, err
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("Error closing database: %v\n", err)
		}
	}()

	// Check if the user exists in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND password = ?", email, password).Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, fmt.Errorf("invalid credentials")
	}

	sessionValues := map[interface{}]interface{}{"email": email}

	err = storage.SetSessionValue(w, r, sessionValues)
	if err != nil {
		return false, fmt.Errorf("failed to set session value: %v", err)
	}
	return true, nil
}

func IsAuthenticated(r *http.Request) bool {
	// Check if the user is authenticated
	session := storage.GetSessionValue(r, "email")
	return session != ""
}

// TODO: Implement password encryption and hashing
func ValidatePassword(letter_password, password_hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(letter_password))
	return err == nil
}

// TODO: Implement password encryption and hashing
func GenerateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
