package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// GetDB initializes and returns a SQLite3 database connection.
func GetDB() (*sql.DB, error) {
	path, err := getDBPath()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// Verify connection.
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// InitDB creates database tables based on the defined models.
func InitDB() error {
	db, err := GetDB()

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Check if the database is already initialized.
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to check if users table exists: %v", err)
	}

	if count == 0 {
		// Create the users table if it doesn't exist.
		err = CreateUserTable(db)
		if err != nil {
			return fmt.Errorf("failed to create users table: %v", err)
		}
	}

	return nil
}

func getDBPath() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		databaseDir := filepath.Join(currentDir, "database")

		// Check if the database directory exists
		if info, err := os.Stat(databaseDir); err == nil && info.IsDir() {
			dbPath := filepath.Join(databaseDir, "timterests.db")
			return dbPath, nil
		}

		parentDir := filepath.Dir(currentDir)

		if parentDir == currentDir {
			return "", fmt.Errorf("could not find 'database' directory in parent tree")
		}

		currentDir = parentDir
	}
}

func CreateUserTable(db *sql.DB) error {

	usersSql := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL
    );`

	_, err := db.Exec(usersSql)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}
	return nil
}
