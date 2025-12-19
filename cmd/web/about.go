package web

import (
	"log"
	"net/http"
	"timterests/internal/storage"
)

// About represents About page content.
type About struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	Body     string `yaml:"body"`
}

// AboutHandler handles requests to the About page and serves it's content.
func AboutHandler(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	var about About

	// Get all articles from the storage
	prefix := "about/"

	aboutFile, err := s.ListS3Objects(r.Context(), prefix)
	if err != nil {
		http.Error(w, "Failed to fetch about info", http.StatusInternalServerError)

		return
	}

	key := *aboutFile[0].Key

	err = s.GetPreparedFile(r.Context(), key, &about)
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
