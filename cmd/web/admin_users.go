package web

import (
	"net/http"

	"github.com/a-h/templ"

	"timterests/internal/auth"
	apperrors "timterests/internal/errors"
)

// AdminUsersPageHandler handles the admin user creation page at /admin/users.
func AdminUsersPageHandler(w http.ResponseWriter, r *http.Request, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	var component templ.Component

	if IsHTMXRequest(r) {
		SetPartialResponseHeaders(w)

		component = AdminUsersDisplay("")
	} else {
		component = AdminUsersPage("")
	}

	err := renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "AdminUsersPageHandler", "render")
	}
}

// CreateUserHandler handles POST /admin/users/create to create a new user.
func CreateUserHandler(w http.ResponseWriter, r *http.Request, a *auth.Auth) {
	if !a.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	if r.Method != http.MethodPost {
		HandleError(w, r, apperrors.MethodNotAllowed(), "CreateUserHandler", "method")

		return
	}

	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if firstName == "" || lastName == "" || email == "" || password == "" {
		renderUserResult(w, r, "All fields are required.", true)

		return
	}

	err := a.CreateUser(r.Context(), firstName, lastName, email, password)
	if err != nil {
		renderUserResult(w, r, "Failed to create user. Please try again.", true)

		return
	}

	renderUserResult(w, r, "User created successfully.", false)
}

func renderUserResult(w http.ResponseWriter, r *http.Request, message string, isError bool) {
	SetPartialResponseHeaders(w)

	status := http.StatusOK
	if isError {
		status = http.StatusUnprocessableEntity
	}

	component := CreateUserResult(message, isError)

	err := renderHTML(w, r, status, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "CreateUserHandler", "renderResult")
	}
}
