package server

import (
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

	s, err := storage.NewS3Storage()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize storage: %v", err))
	}

	NewServer := &Server{
		port:    port,
		storage: s,
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
