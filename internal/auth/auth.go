// Package auth provides authentication and session management functionality.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"timterests/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user in the database with the provided details.
func CreateUser(ctx context.Context, firstName, lastName, email, password string) error {
	db, err := storage.GetDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Printf("Error closing database: %v\n", err)
		}
	}()

	// Generate hashed password
	passwordHash, err := GenerateHash(password)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	// Insert the new user into the database
	_, err = db.ExecContext(
		ctx,
		"INSERT INTO users (first_name, last_name, email, password) VALUES (?, ?, ?, ?)",
		firstName, lastName, email, passwordHash,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Authenticate verifies user credentials and sets session values upon successful authentication.
func Authenticate(ctx context.Context, w http.ResponseWriter, r *http.Request, email, password string) (bool, error) {
	db, err := storage.GetDB(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get database: %w", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Printf("Error closing database: %v\n", err)
		}
	}()

	// Fetch the hashed password for the given email
	var passwordHash string

	err = db.QueryRowContext(ctx, "SELECT password FROM users WHERE email = ?", email).Scan(&passwordHash)
	if err != nil {
		return false, errors.New("invalid credentials")
	}

	// Compare the provided password with the hashed password
	if !ValidatePassword(password, passwordHash) {
		return false, errors.New("invalid credentials")
	}

	sessionValues := map[any]any{"email": email}

	err = SetSessionValue(w, r, sessionValues)
	if err != nil {
		return false, fmt.Errorf("failed to set session value: %w", err)
	}

	return true, nil
}

// IsAuthenticated checks if the user is authenticated based on session values.
func IsAuthenticated(r *http.Request) bool {
	// Check if the user is authenticated
	session := GetSessionValue(r, "email")

	return session != ""
}

// ValidatePassword compares a plaintext password with its hashed version.
func ValidatePassword(password, passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))

	return err == nil
}

// GenerateHash generates a bcrypt hash of the given password.
func GenerateHash(password string) (string, error) {
	if password == "" {
		return "", errors.New("password must not be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}

	return string(hash), nil
}
