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

	"timterests/internal/auth"
	"timterests/internal/storage"
)

// Server provides HTTP server configuration with storage backend.
type Server struct {
	port    int
	Storage *storage.Storage
	auth    *auth.Auth
	oidc    *auth.OIDC
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

	// Initialize Auth. A weak signing key is refused outright: it would let anyone
	// forge a session cookie and bypass sign-in entirely.
	sessionKey := os.Getenv("SESSION_KEY")
	if len(sessionKey) < auth.MinSessionKeyLength {
		panic(fmt.Sprintf(
			"SESSION_KEY must be at least %d characters; it signs session cookies",
			auth.MinSessionKeyLength,
		))
	}

	authInstance := auth.NewAuth(os.Getenv("SESSION_NAME"), sessionKey)

	NewServer := &Server{
		port:    port,
		Storage: store,
		auth:    authInstance,
		oidc:    auth.NewOIDC(auth.OIDCConfigFromEnv(), authInstance),
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
