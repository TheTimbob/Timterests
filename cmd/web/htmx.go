package web

import (
	"net/http"
	"strings"
)

// IsHTMXRequest returns true if the request was initiated by HTMX.
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("Hx-Request") == "true"
}

// SetPartialResponseHeaders prevents the browser from caching HTMX partial responses.
// Without these headers, back-button navigation can serve cached partials without the
// base layout. Cache-Control: no-store forces a re-request; Vary: HX-Request tells
// proxies to treat HTMX and non-HTMX requests as distinct cache entries.
func SetPartialResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

	// Merge "HX-Request" into Vary only if not already present to avoid duplicates.
	existing := w.Header().Get("Vary")

	// Vary: * is a terminal value per RFC 7231 — must not be combined with field-names.
	if strings.TrimSpace(existing) == "*" {
		return
	}

	alreadySet := false

	for token := range strings.SplitSeq(existing, ",") {
		if strings.EqualFold(strings.TrimSpace(token), "HX-Request") {
			alreadySet = true

			break
		}
	}

	if !alreadySet {
		if existing == "" {
			w.Header().Set("Vary", "HX-Request")
		} else {
			w.Header().Set("Vary", existing+", HX-Request")
		}
	}
}
