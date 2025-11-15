package auth

import (
	"fmt"
	"net/http"

	"timterests/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

func CreateUser(firstName, lastName, email, password string) error {
	db, err := storage.GetDB()
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("Error closing database: %v\n", err)
		}
	}()

	// Generate hashed password
	passwordHash, err := GenerateHash(password)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %v", err)
	}

	// Insert the new user into the database
	_, err = db.Exec("INSERT INTO users (first_name, last_name, email, password) VALUES (?, ?, ?, ?)", firstName, lastName, email, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
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

	// Fetch the hashed password for the given email
	var passwordHash string
	err = db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&passwordHash)
	if err != nil {
		return false, fmt.Errorf("invalid credentials")
	}

	// Compare the provided password with the hashed password
	if !ValidatePassword(password, passwordHash) {
		return false, fmt.Errorf("invalid credentials")
	}

	sessionValues := map[any]any{"email": email}

	err = SetSessionValue(w, r, sessionValues)
	if err != nil {
		return false, fmt.Errorf("failed to set session value: %v", err)
	}
	return true, nil
}

func IsAuthenticated(r *http.Request) bool {
	// Check if the user is authenticated
	session := GetSessionValue(r, "email")
	return session != ""
}
func ValidatePassword(password, password_hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))
	return err == nil
}

func GenerateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
