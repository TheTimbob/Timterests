package web

import (
	"context"
	"log"
	"net/http"
	"timterests/internal/storage"
)

type About struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	Body     string `yaml:"body"`
}

func AboutHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage) {
	var about About

	// Get all articles from the storage
	prefix := "about/"
	aboutFile, err := storage.ListObjects(context.Background(), storageInstance, prefix)
	if err != nil {
		http.Error(w, "Failed to fetch about info", http.StatusInternalServerError)
		return
	}

	key := *aboutFile[0].Key
	err = storage.GetPreparedFile(key, &about, storageInstance)
	if err != nil {
		http.Error(w, "Failed to prepare about info", http.StatusInternalServerError)
		log.Printf("Error fetching about info: %v", err)
		return
	}

	component := AboutForm(about)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in AboutHandler: %e", err)
		return
	}
}
