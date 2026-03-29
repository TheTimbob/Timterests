package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	// Import godotenv for automatic .env file loading.
	_ "github.com/joho/godotenv/autoload"

	"timterests/internal/auth"
	apperrors "timterests/internal/errors"
	"timterests/internal/storage"
)

// Server provides HTTP server configuration with storage backend.
type Server struct {
	port    int
	storage *storage.Storage
	auth    *auth.Auth
}

// NewServer creates and configures a new HTTP server instance.
func NewServer() *http.Server {
	// Initialize structured error logger (best-effort; console logging still works if this fails).
	logPath := os.Getenv("ERROR_LOG_PATH")
	if logPath == "" {
		logPath = "logs/errors.log"
	}

	if err := apperrors.InitLogger(logPath); err != nil {
		log.Printf("warning: could not initialize error log file (%v); falling back to console only", err)
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(fmt.Sprintf("failed to parse PORT: %v", err))
	}

	// Initialize Storage (handles both S3 and local)
	store, err := storage.NewStorage(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to initialize storage: %v", err))
	}

	// Initialize Auth with session name from environment
	authInstance := auth.NewAuth(os.Getenv("SESSION_NAME"))

	NewServer := &Server{
		port:    port,
		storage: store,
		auth:    authInstance,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Ensure the error log file is closed during graceful shutdown.
	server.RegisterOnShutdown(apperrors.CloseLogger)

	return server
}
