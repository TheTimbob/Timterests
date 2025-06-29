package web

import (
	"log"
	"net/http"
	"timterests/internal/storage"
)

func HomeHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage) {

	latestArticle, err := GetLatestArticle(storageInstance)
	if err != nil {
		http.Error(w, "Failed to fetch latest article", http.StatusInternalServerError)
		log.Printf("Error fetching latest article: %v", err)
		return
	}

	featuredProject, err := GetFeaturedProject(storageInstance)
	if err != nil {
		http.Error(w, "Failed to fetch featured project", http.StatusInternalServerError)
		log.Printf("Error fetching featured project: %v", err)
		return
	}

	component := HomeForm(latestArticle, featuredProject)

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering home page: %v", err)
	}
}
