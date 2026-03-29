package web

import (
	"log"
	"net/http"
	"timterests/internal/service"
	"timterests/internal/storage"
)

// HomeHandler handles requests to the home page.
func HomeHandler(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	latestArticle, err := service.GetLatestArticle(r.Context(), s)
	if err != nil {
		http.Error(w, "Failed to fetch latest article", http.StatusInternalServerError)
		log.Printf("Error fetching latest article: %v", err)

		return
	}

	featuredProject, err := service.GetFeaturedProject(r.Context(), s, "Timterests")
	if err != nil {
		http.Error(w, "Failed to fetch featured project", http.StatusInternalServerError)
		log.Printf("Error fetching featured project: %v", err)

		return
	}

	component := HomeForm(latestArticle, featuredProject)

	err = component.Render(r.Context(), w)
	if err != nil {
		log.Printf("HomeHandler: failed to render: %v", err)
	}
}
