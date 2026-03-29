package web

import (
	"log"
	"net/http"
	"timterests/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// About represents About page content.
type About struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	Body     string `yaml:"body"`
}

// AboutHandler handles requests to the About page and serves its content.
func AboutHandler(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	var about About

	// Get all articles from the storage
	prefix := "about/"

	aboutFile, err := s.ListObjects(r.Context(), prefix)
	if err != nil {
		http.Error(w, "Failed to fetch about info", http.StatusInternalServerError)

		return
	}

	if len(aboutFile) == 0 {
		http.Error(w, "Not Found", http.StatusNotFound)

		return
	}

	key := aws.ToString(aboutFile[0].Key)
	if key == "" {
		http.Error(w, "Not Found", http.StatusNotFound)

		return
	}

	err = s.GetPreparedFile(r.Context(), key, &about)
	if err != nil {
		http.Error(w, "Failed to prepare about info", http.StatusInternalServerError)
		log.Printf("Error fetching about info: %v", err)

		return
	}

	component := AboutForm(about)

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		log.Printf("AboutHandler: failed to render: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
