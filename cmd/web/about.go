package web

import (
	"context"
	"log"
	"net/http"
	"path"
	"timterests/internal/models"
	"timterests/internal/storage"
)

func AboutHandler(w http.ResponseWriter, r *http.Request, storageInstance models.Storage) {
	var about models.About

	// Get all articles from the storage
	prefix := "about/"
	aboutFile, err := storage.ListObjects(context.Background(), storageInstance, prefix)
	if err != nil {
		http.Error(w, "Failed to fetch about info", http.StatusInternalServerError)
		return
	}

	key := *aboutFile[0].Key
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	document, err := storage.ReadFile(key, localFilePath, storageInstance)
	if err != nil {
		http.Error(w, "Failed to read about file", http.StatusInternalServerError)
		return
	}

	about = models.About{
		Title:    document.Title,
		Subtitle: document.Subtitle,
		Body:     document.Body,
	}

	component := AboutForm(about)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in AboutHandler: %e", err)
	}
}
