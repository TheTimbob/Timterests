package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"internal/templates"
)

func main() {
	component := templates.hello("Tim")
    
    // Runs server
	http.Handle("/", templ.Handler(component))

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
