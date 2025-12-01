package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// NewSQLiteDatabase initializes and returns a SQLite3 database connection.
func NewSQLiteDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
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
func InitDB(path string) error {
	db, err := NewSQLiteDatabase(path)

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
		err = CreatUserTable(db)
		if err != nil {
			return fmt.Errorf("failed to create users table: %v", err)
		}
	}

	return nil
}

func GetDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "database/timterests.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	return db, nil
}

func CreatUserTable(db *sql.DB) error {

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
