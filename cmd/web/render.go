package web

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
)

// renderHTML renders a templ component into a buffer, then writes the response with the given
// status code and Content-Type header. It returns an error only if rendering fails (before any
// headers are written), so callers can still send an error response. Write failures are logged
// but not returned because headers have already been sent at that point.
func renderHTML(w http.ResponseWriter, r *http.Request, status int, component templ.Component) error {
	buf := &bytes.Buffer{}

	err := component.Render(r.Context(), buf)
	if err != nil {
		return fmt.Errorf("rendering component: %w", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	_, err = buf.WriteTo(w)
	if err != nil {
		log.Printf("renderHTML: failed to write response: %v", err)
	}

	return nil
}
