package web

import (
	"net/http"
	"strings"

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

	var key string
	for _, obj := range aboutFile {
		k := aws.ToString(obj.Key)
		if strings.HasSuffix(k, ".yaml") {
			key = k
			break
		}
	}

	if key == "" {
		HandleError(w, r, apperrors.NotFound(nil), "AboutHandler", "getKey")

		return
	}

	err = s.GetPreparedFile(r.Context(), key, &about)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "AboutHandler", "getPreparedFile")

		return
	}

	body, err := s.GetDocumentBody(r.Context(), key)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "AboutHandler", "getDocumentBody")

		return
	}

	about.Body = body
	about.GitHub = strings.TrimSpace(about.GitHub)
	about.Email = strings.TrimSpace(about.Email)

	component := AboutForm(about)

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "AboutHandler", "render")
	}
}
