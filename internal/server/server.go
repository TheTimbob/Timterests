package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"timterests/internal/models"
	"timterests/internal/storage"
)

type Server struct {
	port int

	storage *models.Storage
}

func NewServer() *http.Server {

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	storageInstance, err := storage.NewStorage()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize storage: %v", err))
	}

	NewServer := &Server{
		port:    port,
		storage: storageInstance,
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
