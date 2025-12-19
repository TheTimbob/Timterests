package auth_test

import (
	"testing"

	"timterests/internal/auth"
)

func TestGenerateHash(t *testing.T) {
	t.Parallel()

	password := "Password1234!"

	hash, err := auth.GenerateHash(password)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash == "" {
		t.Errorf("Expected non-empty hash")
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	password := "Password1234!"

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
}
