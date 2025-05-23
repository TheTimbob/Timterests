package web

import (
	"context"
	"log"
	"net/http"
	"path"
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
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	file, err := storage.GetFile(key, localFilePath, storageInstance)
	if err != nil {
		http.Error(w, "Failed to read about file", http.StatusInternalServerError)
		return
	}

	if err := storage.DecodeFile(file, &about); err != nil {
		log.Println("Failed to decode file:", err)
		return
	}

	body, err := storage.BodyToHTML(about.Body)
	if err != nil {
		return
	}

	about.Body = body

	component := AboutForm(about)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in AboutHandler: %e", err)
	}
}
