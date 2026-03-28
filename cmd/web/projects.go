package web

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"timterests/internal/auth"
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
		message := "Failed to fetch projects"
		http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusInternalServerError)

		return
	}

	for i := range projects {
		projects[i].Body = storage.RemoveHTMLTags(projects[i].Body)
		v := reflect.ValueOf(projects[i])
		tags = storage.GetTags(v, tags)
	}

	if IsHTMXRequest(r) {
		SetPartialResponseHeaders(w)
		component = ProjectsList(projects, design)
	} else {
		component = ProjectsListPage(projects, tags, design)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in ProjectPosts: %e", err)
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
			var component templ.Component

			authenticated := a.IsAuthenticated(r)

			if IsHTMXRequest(r) {
				SetPartialResponseHeaders(w)
				component = ProjectDisplay(project, authenticated)
			} else {
				component = ProjectPage(project, authenticated)
			}

			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Printf("Error rendering in GetProjectsHandler: %e", err)
			}
		}
	}
}
