package web

import (
	"net/http"

	apperrors "timterests/internal/errors"

	"timterests/internal/auth"

	"github.com/a-h/templ"
)

func AdminPageHandler(w http.ResponseWriter, r *http.Request, a *auth.Auth) {
	var component templ.Component

	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	component = AdminPage()

	err := renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "AdminPageHandler", "render")
	}
}
