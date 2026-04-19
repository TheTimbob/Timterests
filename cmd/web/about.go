package web

import (
	"net/http"

	apperrors "timterests/internal/errors"
	"timterests/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type About struct {
	Title      string `yaml:"title"`
	Subtitle   string `yaml:"subtitle"`
	Body       string `yaml:"body"`
	Name       string `yaml:"name"`
	Specialty  string `yaml:"specialty"`
	Location   string `yaml:"location"`
	GitHub     string `yaml:"github"`
	Email      string `yaml:"email"`
}

func AboutHandler(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	var about About

	prefix := "about/"

	aboutFile, err := s.ListObjects(r.Context(), prefix)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "AboutHandler", "listObjects")

		return
	}

	if len(aboutFile) == 0 {
		HandleError(w, r, apperrors.NotFound(nil), "AboutHandler", "findAbout")

		return
	}

	key := aws.ToString(aboutFile[0].Key)
	if key == "" {
		HandleError(w, r, apperrors.NotFound(nil), "AboutHandler", "getKey")

		return
	}

	err = s.GetPreparedFile(r.Context(), key, &about)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "AboutHandler", "getPreparedFile")

		return
	}

	component := AboutForm(about)

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "AboutHandler", "render")
	}
}
