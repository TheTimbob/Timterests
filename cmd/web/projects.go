package web

import (
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
		log.Printf("ProjectsPageHandler: failed to fetch projects: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	for i := range projects {
		projects[i].Body = storage.RemoveHTMLTags(projects[i].Body)
		v := reflect.ValueOf(projects[i])
		tags = storage.GetTags(v, tags)
	}

	if r.Header.Get("Hx-Request") == "true" {
		component = ProjectsList(projects, design)
	} else {
		component = ProjectsListPage(projects, tags, design)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		log.Printf("ProjectsPageHandler: failed to render: %v", err)
	}
}

// GetProjectHandler handles requests to get a specific project by its ID.
func GetProjectHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, projectID string, a *auth.Auth) {
	projects, err := service.ListProjects(r.Context(), s, "all")
	if err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)

		return
	}

	found := false

	for _, project := range projects {
		if project.ID == projectID {
			found = true

			var component templ.Component

			authenticated := a.IsAuthenticated(r)

			if r.Header.Get("Hx-Request") == "true" {
				component = ProjectDisplay(project, authenticated)
			} else {
				component = ProjectPage(project, authenticated)
			}

			err = component.Render(r.Context(), w)
			if err != nil {
				log.Printf("GetProjectHandler: failed to render: %v", err)
			}
		}
	}

	if !found {
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}
