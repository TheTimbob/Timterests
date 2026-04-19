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

func ReadingListPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, currentTag, design string) {
	var (
		component templ.Component
		tags      []string
	)

	books, err := service.ListBooks(r.Context(), s, currentTag)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "ReadingListPageHandler", "listBooks")

		return
	}

	for i := range books {
		v := reflect.ValueOf(books[i])
		tags = storage.GetTags(v, tags)
	}

	if r.Header.Get("Hx-Request") == "true" {
		component = ReadingListList(books, design)
	} else {
		component = ReadingListPage(books, tags, design)
	}

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "ReadingListPageHandler", "render")
	}
}

func GetReadingListBook(w http.ResponseWriter, r *http.Request, s storage.Storage, bookID string, a *auth.Auth) {
	books, err := service.ListBooks(r.Context(), s, "all")
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "GetReadingListBook", "listBooks")

		return
	}

	for _, book := range books {
		if book.ID == bookID {
			body, err := s.GetDocumentBody(r.Context(), book.S3Key)
			if err != nil {
				HandleError(w, r, apperrors.NotFound(err), "GetReadingListBook", "getBody")

				return
			}

			dc := model.DisplayContent{
				ID:    book.ID,
				S3Key: book.S3Key,
				Body:  body,
			}

			var component templ.Component

			authenticated := a.IsAuthenticated(r)

			if r.Header.Get("Hx-Request") == "true" {
				component = BookDisplay(book, dc, authenticated)
			} else {
				component = BookPage(book, dc, authenticated)
			}

			err = renderHTML(w, r, http.StatusOK, component)
			if err != nil {
				HandleError(w, r, apperrors.RenderFailed(err), "GetReadingListBook", "render")

				return
			}

			return
		}
	}

	HandleError(w, r, apperrors.NotFound(nil), "GetReadingListBook", "findBook")
}
