// Package main provides the API server entry point for the timterests application.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"timterests/internal/server"
	"timterests/internal/storage"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := apiServer.Shutdown(ctx)
	if err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	// Initialize the database
	err := storage.InitDB(context.Background())
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
	}

	// Initialize the server
	server := server.NewServer()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")

	tlsStarted := false

	_, err = os.Stat(certFile) // #nosec G703 -- certFile comes from environment variable
	if err == nil {
		_, err := os.Stat(keyFile) // #nosec G703 -- keyFile comes from environment variable
		if err == nil {
			err := server.ListenAndServeTLS(certFile, keyFile)
			if err != nil {
				log.Fatalf("Failed to start server with TLS: %v", err)
			}

			tlsStarted = true
		}
	}

	if !tlsStarted {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
