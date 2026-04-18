package web

import (
	"log"
	"net/http"
	"reflect"
	"timterests/internal/auth"
	"timterests/internal/model"
	"timterests/internal/service"
	"timterests/internal/storage"

	"github.com/a-h/templ"
)

// ProjectsPageHandler handles requests to the projects page,
// ensuring authentication and rendering the appropriate content.
func ProjectsPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, currentTag, design string) {
	var (
		component templ.Component
		tags      []string
	)

	projects, err := service.ListProjects(r.Context(), s, currentTag)
	if err != nil {
		log.Printf("ProjectsPageHandler: failed to fetch projects: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	for i := range projects {
		v := reflect.ValueOf(projects[i])
		tags = storage.GetTags(v, tags)
	}

	if r.Header.Get("Hx-Request") == "true" {
		component = ProjectsList(projects, design)
	} else {
		component = ProjectsListPage(projects, tags, design)
	}

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		log.Printf("ProjectsPageHandler: failed to render: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetProjectHandler handles requests to get a specific project by its ID.
func GetProjectHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, projectID string, a *auth.Auth) {
	projects, err := service.ListProjects(r.Context(), s, "all")
	if err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)

		return
	}

	for _, project := range projects {
		if project.ID == projectID {
			body, err := s.GetDocumentBody(r.Context(), project.S3Key)
			if err != nil {
				log.Printf("GetProjectHandler: failed to load body for %s: %v", project.S3Key, err)
				http.Error(w, "Not Found", http.StatusNotFound)

				return
			}

			dc := model.DisplayContent{
				ID:    project.ID,
				S3Key: project.S3Key,
				Body:  body,
			}

			var component templ.Component

			authenticated := a.IsAuthenticated(r)

			if r.Header.Get("Hx-Request") == "true" {
				component = ProjectDisplay(dc, project.Repository, authenticated)
			} else {
				component = ProjectPage(dc, project.Repository, authenticated)
			}

			err = renderHTML(w, r, http.StatusOK, component)
			if err != nil {
				log.Printf("GetProjectHandler: failed to render: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			return
		}
	}

	http.Error(w, "Not Found", http.StatusNotFound)
}
