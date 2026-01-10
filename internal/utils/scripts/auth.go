// Package scripts provides utility scripts for administrative tasks.
package scripts

import (
	"context"
	"fmt"
	"os"

	"timterests/internal/auth"
	"timterests/internal/storage"
)

// CreateUser creates a new user in the database with the provided credentials.
func CreateUser(firstName, lastName, email, password string) error {
	ctx := context.Background()

	err := storage.InitDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create auth instance with session name from environment
	a := auth.NewAuth(os.Getenv("SESSION_NAME"))

	err = a.CreateUser(ctx, firstName, lastName, email, password)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
