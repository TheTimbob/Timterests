package web

import (
	"net/http"

	apperrors "timterests/internal/errors"
)

// HandleError logs the error with structured formatting and renders an HTML error page.
// It classifies any error into an AppError, enriches it with handler context, logs it,
// and renders the appropriate error page for the status code.
func HandleError(w http.ResponseWriter, r *http.Request, err error, handler, action string) {
	appErr := apperrors.Classify(err)
	appErr = appErr.WithHandler(handler, action)
	apperrors.LogError(appErr)

	component := ErrorPage(appErr.HTTPStatus, appErr.Message)

	renderErr := renderHTML(w, r, appErr.HTTPStatus, component)
	if renderErr != nil {
		http.Error(w, appErr.Message, appErr.HTTPStatus)
	}
}
