package web

import (
	"net/http"

	apperrors "timterests/internal/errors"
	"timterests/internal/service"
	"timterests/internal/storage"
)

func HomeHandler(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	latestArticle, err := service.GetLatestArticle(r.Context(), s)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "HomeHandler", "getLatestArticle")

		return
	}

	featuredProject, err := service.GetFeaturedProject(r.Context(), s, "Timterests")
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "HomeHandler", "getFeaturedProject")

		return
	}

	component := HomeForm(latestArticle, featuredProject)

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "HomeHandler", "render")
	}
}
