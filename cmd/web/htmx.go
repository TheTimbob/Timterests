package web

import "net/http"

// IsHTMXRequest returns true if the request was initiated by HTMX.
// HTMX sets the "HX-Request: true" header on all requests it initiates.
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// IsBackButtonNavigation returns true if this request appears to be from browser
// back/forward navigation, direct URL entry, or any non-HTMX request to a route
// that normally serves partial content.
//
// These requests should always receive a full-page response (base layout + content)
// rather than the partial HTMX swap. Without this check, back-button navigation
// can render cached partial HTML without the base layout, styles, or navigation.
//
// Logic:
//   - HTMX requests include "HX-Request: true" header
//   - Back-button, direct nav, and bookmark visits do NOT include this header
//   - Therefore: no HX-Request header → treat as back/direct nav → return full page
func IsBackButtonNavigation(r *http.Request) bool {
	return !IsHTMXRequest(r)
}

// SetPartialResponseHeaders applies cache-prevention headers to HTMX partial responses.
//
// Why this matters: HTMX swaps content in-place and pushes the URL to browser history
// via hx-push-url. The browser caches the partial HTML at that URL. When the user
// presses Back, the browser serves the cached partial — bypassing the server entirely —
// resulting in a page with no base layout, styles, or navigation.
//
// Setting Cache-Control: no-store forces the browser to re-request from the server
// on back-button navigation. Since back-button requests don't include "HX-Request: true",
// the server correctly returns the full page.
//
// Vary: HX-Request tells caches (CDN, proxy) to treat HTMX and non-HTMX requests
// as distinct cache entries, preventing a partial response from being served to a
// full-page request.
func SetPartialResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Vary", "HX-Request")
}
