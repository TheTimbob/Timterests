package auth_test

import (
	"testing"

	"timterests/internal/auth"
)

func TestAuthPassword(t *testing.T) {
	t.Parallel()

	password := "Password1234!"

	t.Run("generate Hash", func(t *testing.T) {
		t.Parallel()

		hash, err := auth.GenerateHash(password)
		if err != nil {
			t.Fatalf("Failed to generate hash: %v", err)
		}

		if hash == "" {
			t.Errorf("Expected non-empty hash")
		}
	})

	t.Run("validate password is correct", func(t *testing.T) {
		t.Parallel()

		hash, err := auth.GenerateHash(password)
		if err != nil {
			t.Fatalf("Failed to generate hash: %v", err)
		}

		if !auth.ValidatePassword(password, hash) {
			t.Errorf("Expected password to be valid")
		}

		if auth.ValidatePassword("wrongpassword", hash) {
			t.Errorf("Expected password to be invalid")
		}
	})

	t.Run("generate Hash with empty password", func(t *testing.T) {
		t.Parallel()

		_, err := auth.GenerateHash("")
		if err == nil {
			t.Errorf("Expected error for empty password")
		}
	})
}
