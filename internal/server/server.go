package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	// Import godotenv for automatic .env file loading.
	_ "github.com/joho/godotenv/autoload"

	"timterests/internal/storage"
)

// Server provides HTTP server configuration with storage backend.
type Server struct {
	port    int
	storage *storage.Storage
}

// NewServer creates and configures a new HTTP server instance.
func NewServer() *http.Server {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(fmt.Sprintf("failed to parse PORT: %v", err))
	}

	// Initialize Storage (handles both S3 and local)
	store, err := storage.NewStorage(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to initialize storage: %v", err))
	}

	NewServer := &Server{
		port:    port,
		storage: store,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
