package web

import (
	"fmt"
	"net/http"
	"time"

	"timterests/internal/storage"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    // Check if the request method is POST
    if r.Method == http.MethodPost {
        email := r.FormValue("email")
        password := r.FormValue("password")

        // Perform authentication
        isAuthenticated, err := Authenticate(w, email, password)
        if isAuthenticated && err == nil {
            http.Redirect(w, r, "/letters", http.StatusSeeOther)
            return
        } else {
            http.Error(w, "Authentication: "+err.Error(), http.StatusInternalServerError)
            return
        }
    }

    // Render the login page for the initial load
    component := LoginPage()
    if r.Header.Get("HX-Request") == "true" {
        // Render the login container for main inner html replacement
        component = LoginContainer()
    }

    err := component.Render(r.Context(), w)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
}

func Authenticate(w http.ResponseWriter, email, password string) (bool, error) {

    db, err := storage.GetDB()
    if err != nil {
        return false, err
    }
    defer db.Close()

    // Check if the user exists in the database
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND password = ?", email, password).Scan(&count)
    if err != nil {
        return false, err
    }

    if count == 0 {
        return false, fmt.Errorf("invalid credentials")
    }

    // On successful authentication, set cookie "session" with email.
    cookie := &http.Cookie{
        Name:     "session",
        Value:    email,
        Expires:  time.Now().Add(24 * time.Hour),
        HttpOnly: true,
    }
    http.SetCookie(w, cookie)
    return true, nil
}
