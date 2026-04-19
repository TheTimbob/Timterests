package server_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupHealthTestDB(t *testing.T) {
	t.Helper()

	dbDir := filepath.Join(t.TempDir(), "database")

	err := os.MkdirAll(dbDir, 0750)
	if err != nil {
		t.Fatalf("failed to create database dir: %v", err)
	}

	dbPath := filepath.Join(dbDir, "timterests.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	defer func() {
		_ = db.Close()
	}()

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	t.Chdir(filepath.Dir(dbDir))
}
