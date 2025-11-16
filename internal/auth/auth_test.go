package auth

import "testing"

func TestGenerateHash(t *testing.T) {
	password := "Password1234!"
	hash, err := GenerateHash(password)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash == "" {
		t.Errorf("Expected non-empty hash")
	}
}

func TestValidatePassword(t *testing.T) {
	password := "Password1234!"
	hash, err := GenerateHash(password)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if !ValidatePassword(password, hash) {
		t.Errorf("Expected password to be valid")
	}

	if ValidatePassword("wrongpassword", hash) {
		t.Errorf("Expected password to be invalid")
	}
}
