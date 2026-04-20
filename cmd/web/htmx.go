package web

import (
	"net/http"
	"strings"
)

// IsHTMXRequest returns true if the request was initiated by HTMX.
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("Hx-Request") == "true"
}

// SetVaryHeader merges "HX-Request" into the Vary header if not already present.
// Must be set on ALL responses from routes that serve both full-page and partial
// variants, so intermediary caches key on the header and never serve the wrong variant.
func SetVaryHeader(w http.ResponseWriter) {
	existing := w.Header().Get("Vary")

	// Vary: * is a terminal value per RFC 7231 — must not be combined with field-names.
	if strings.TrimSpace(existing) == "*" {
		return
	}

	for token := range strings.SplitSeq(existing, ",") {
		if strings.EqualFold(strings.TrimSpace(token), "HX-Request") {
			return
		}
	}

	if existing == "" {
		w.Header().Set("Vary", "HX-Request")
	} else {
		w.Header().Set("Vary", existing+", HX-Request")
	}
}

// SetPartialResponseHeaders prevents the browser from caching HTMX partial responses.
// Cache-Control: no-store forces a re-request on back-button navigation so the server
// returns a full page. Vary: HX-Request is set via SetVaryHeader so all response variants
// from the same URL are correctly keyed by intermediary caches.
func SetPartialResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	SetVaryHeader(w)
}
