// Package storage provides database and S3 storage operations for the application.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	// Import SQLite3 driver for database/sql.
	_ "github.com/mattn/go-sqlite3"
)

// GetDB initializes and returns a SQLite3 database connection.
func GetDB(ctx context.Context) (*sql.DB, error) {
	path, err := getDBPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	// Verify connection.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// InitDB creates database tables based on the defined models.
func InitDB(ctx context.Context) error {
	db, err := GetDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Check if the database is already initialized.
	var count int

	err = db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'",
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if users table exists: %w", err)
	}

	if count == 0 {
		// Create the users table if it doesn't exist.
		err = CreateUserTable(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to create users table: %w", err)
		}
	}

	return nil
}

// getDBPath locates the database file by searching for the 'database' directory
// in parent directories.
func getDBPath() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		databaseDir := filepath.Join(currentDir, "database")

		// Check if the database directory exists
		info, err := os.Stat(databaseDir)
		if err == nil && info.IsDir() {
			dbPath := filepath.Join(databaseDir, "timterests.db")

			return dbPath, nil
		}

		parentDir := filepath.Dir(currentDir)

		if parentDir == currentDir {
			return "", errors.New("could not find 'database' directory in parent tree")
		}

		currentDir = parentDir
	}
}

// CreateUserTable creates the 'users' table in the database.
func CreateUserTable(ctx context.Context, db *sql.DB) error {
	usersSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL
    );`

	_, err := db.ExecContext(ctx, usersSQL)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}
