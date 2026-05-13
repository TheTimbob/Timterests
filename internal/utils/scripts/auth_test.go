package scripts_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"timterests/internal/utils/scripts"

	_ "github.com/mattn/go-sqlite3"
)

func TestCreateUser(t *testing.T) {
	t.Run("creates user in database", func(t *testing.T) {
		dbDir := filepath.Join(t.TempDir(), "database")

		err := os.MkdirAll(dbDir, 0750)
		if err != nil {
			t.Fatalf("failed to create database dir: %v", err)
		}

		t.Chdir(filepath.Dir(dbDir))

		err = scripts.CreateUser("Jane", "Doe", "jane@example.com", "SecurePass123!")
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		db, err := sql.Open("sqlite3", filepath.Join(dbDir, "timterests.db"))
		if err != nil {
			t.Fatalf("failed to open db: %v", err)
		}
		defer db.Close()

		var email string

		err = db.QueryRowContext(
			context.Background(),
			"SELECT email FROM users WHERE email = ?",
			"jane@example.com",
		).Scan(&email)
		if err != nil {
			t.Fatalf("failed to query user: %v", err)
		}

		if email != "jane@example.com" {
			t.Errorf("expected email jane@example.com, got %q", email)
		}
	})

	t.Run("returns error for duplicate email", func(t *testing.T) {
		dbDir := filepath.Join(t.TempDir(), "database")

		err := os.MkdirAll(dbDir, 0750)
		if err != nil {
			t.Fatalf("failed to create database dir: %v", err)
		}

		t.Chdir(filepath.Dir(dbDir))

		err = scripts.CreateUser("First", "User", "dup@example.com", "Pass123!")
		if err != nil {
			t.Fatalf("first CreateUser failed: %v", err)
		}

		err = scripts.CreateUser("Second", "User", "dup@example.com", "Pass456!")
		if err == nil {
			t.Error("expected error for duplicate email")
		}
	})
}
