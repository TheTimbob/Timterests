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

func LettersPageHandler(
	w http.ResponseWriter,
	r *http.Request,
	s storage.Storage,
	currentTag, design string,
	a *auth.Auth) {
	var (
		component templ.Component
		tags      []string
	)

	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	letters, err := service.ListLetters(r.Context(), s, currentTag)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "LettersPageHandler", "listLetters")

		return
	}

	for i := range letters {
		v := reflect.ValueOf(letters[i])
		tags = storage.GetTags(v, tags)
	}

	if IsHTMXRequest(r) {
		SetPartialResponseHeaders(w)

		component = LettersList(letters, design)
	} else {
		component = LettersListPage(letters, tags, design)
	}

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "LettersPageHandler", "render")
	}
}

func GetLetterHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, letterID string, a *auth.Auth) {
	authenticated := a.IsAuthenticated(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	letters, err := service.ListLetters(r.Context(), s, "all")
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "GetLetterHandler", "listLetters")

		return
	}

	var component templ.Component

	for _, letter := range letters {
		if letter.ID == letterID {
			body, err := s.GetDocumentBody(r.Context(), letter.S3Key)
			if err != nil {
				HandleError(w, r, apperrors.NotFound(err), "GetLetterHandler", "getBody")

				return
			}

			dc := model.DisplayContent{
				ID:    letter.ID,
				S3Key: letter.S3Key,
				Body:  body,
			}

			if IsHTMXRequest(r) {
				SetPartialResponseHeaders(w)

				component = LetterDisplay(dc, authenticated)
			} else {
				component = LetterPage(dc, authenticated)
			}

			err = renderHTML(w, r, http.StatusOK, component)
			if err != nil {
				HandleError(w, r, apperrors.RenderFailed(err), "GetLetterHandler", "render")

				return
			}

			return
		}
	}

	HandleError(w, r, apperrors.NotFound(nil), "GetLetterHandler", "findLetter")
}
