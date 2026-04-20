package web

import (
	"net/http"
	"reflect"

	apperrors "timterests/internal/errors"

	"timterests/internal/auth"
	"timterests/internal/model"
	"timterests/internal/service"
	"timterests/internal/storage"

	"github.com/a-h/templ"
)

func ProjectsPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, currentTag, design string) {
	var (
		component templ.Component
		tags      []string
	)

	projects, err := service.ListProjects(r.Context(), s, currentTag)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "ProjectsPageHandler", "listProjects")

		return
	}

	for i := range projects {
		v := reflect.ValueOf(projects[i])
		tags = storage.GetTags(v, tags)
	}

	if IsHTMXRequest(r) {
		SetPartialResponseHeaders(w)

		component = ProjectsList(projects, design)
	} else {
		component = ProjectsListPage(projects, tags, design)
	}

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "ProjectsPageHandler", "render")
	}
}

func GetProjectHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, projectID string, a *auth.Auth) {
	projects, err := service.ListProjects(r.Context(), s, "all")
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "GetProjectHandler", "listProjects")

		return
	}

	for _, project := range projects {
		if project.ID == projectID {
			body, err := s.GetDocumentBody(r.Context(), project.S3Key)
			if err != nil {
				HandleError(w, r, apperrors.NotFound(err), "GetProjectHandler", "getBody")

				return
			}

			dc := model.DisplayContent{
				ID:    project.ID,
				S3Key: project.S3Key,
				Body:  body,
			}

			var component templ.Component

			authenticated := a.IsAuthenticated(r)

			if IsHTMXRequest(r) {
				SetPartialResponseHeaders(w)

				component = ProjectDisplay(dc, project.Repository, authenticated)
			} else {
				component = ProjectPage(dc, project.Repository, authenticated)
			}

			err = renderHTML(w, r, http.StatusOK, component)
			if err != nil {
				HandleError(w, r, apperrors.RenderFailed(err), "GetProjectHandler", "render")

				return
			}

			return
		}
	}

	HandleError(w, r, apperrors.NotFound(nil), "GetProjectHandler", "findProject")
}
