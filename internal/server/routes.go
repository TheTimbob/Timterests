package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"timterests/cmd/web"
	"timterests/internal/auth"
	apperrors "timterests/internal/errors"
)

// Asset cache lifetimes. Filenames are not content-hashed, so changes only
// propagate once the window lapses.
//
// Images and fonts are the heavy assets and rarely change, so they get a long
// window. CSS and JS are small and edited often, and a long window there just
// means staring at a stale stylesheet wondering why a fix did not land — so they
// get barely any.
const (
	longCacheControl    = "public, max-age=604800"
	defaultCacheControl = "public, max-age=60"
)

// RegisterRoutes configures all HTTP routes and returns the handler.
func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Favicon Route
	mux.Handle("/favicon.ico", http.FileServer(http.Dir(".")))

	// Serve static files from the "storage" directory
	mux.Handle("/storage/", http.StripPrefix("/storage/", http.FileServer(http.Dir("storage"))))

	// Serve static files from the "web" directory
	fileServer := http.FileServer(http.FS(web.Files))
	mux.Handle("/assets/", staticCacheMiddleware(fileServer))

	// Home Routes
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.HomeHandler(w, r, *s.Storage)
	}))
	mux.Handle("/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.HomeHandler(w, r, *s.Storage)
	}))
	mux.Handle("/web", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.HomeHandler(w, r, *s.Storage)
	}))
	mux.Handle("/web/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.HomeHandler(w, r, *s.Storage)
	}))

	mux.Handle("/admin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.AdminPageHandler(w, r, s.auth)
	}))

	mux.Handle("/admin/documents", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.AdminDocumentsPageHandler(w, r, *s.Storage, s.auth)
	}))

	mux.Handle("/admin/documents/delete", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.DeleteDocumentHandler(w, r, *s.Storage, s.auth)
	}))

	mux.Handle("/admin/upload", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			web.UploadDocumentHandler(w, r, *s.Storage, s.auth)

			return
		}

		web.UploadPageHandler(w, r, s.auth)
	}))

	mux.Handle("/writer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			docType, key string
			typeID       int
		)

		err := r.ParseForm()
		if err != nil {
			web.HandleError(w, r, apperrors.ParseFormFailed(err), "writer", "parseForm")

			return
		}

		docType = r.FormValue("document-type")
		if docType == "" {
			docType = "articles" // default
		}

		typeIDString := r.FormValue("type-id")
		if typeIDString != "" {
			var err error

			typeID, err = strconv.Atoi(typeIDString)
			if err != nil {
				web.HandleError(w, r, apperrors.BadRequest(err), "writer", "parseTypeID")

				return
			}
		}

		key = r.FormValue("document-key")

		web.WriterPageHandler(w, r, *s.Storage, docType, key, typeID, s.auth)
	}))

	mux.Handle("/write", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.WriteDocumentHandler(w, r, *s.Storage, s.auth)
	}))

	mux.Handle("/download", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		documentKey := r.URL.Query().Get("key")
		web.DownloadDocumentHandler(w, r, *s.Storage, documentKey, s.auth)
	}))

	mux.Handle("/download/new", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.DownloadNewDocumentHandler(w, r, s.auth)
	}))

	// SEO
	mux.HandleFunc("/robots.txt", web.RobotsHandler)
	mux.Handle("/sitemap.xml", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.SitemapHandler(w, r, *s.Storage)
	}))
	mux.Handle("/rss.xml", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.RSSHandler(w, r, *s.Storage)
	}))

	// Health check
	mux.HandleFunc("/health", s.HealthHandler)

	// About Routes
	mux.Handle("/about", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.AboutHandler(w, r, *s.Storage)
	}))

	// Login Routes
	mux.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.LoginHandler(w, r)
	}))
	mux.Handle("/auth/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.OIDCLoginHandler(w, r, s.oidc)
	}))
	mux.Handle("/auth/callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.OIDCCallbackHandler(w, r, s.oidc)
	}))
	mux.Handle("/logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.LogoutHandler(w, r, s.auth, s.oidc)
	}))

	// Article Routes
	mux.Handle("/articles", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		design := r.URL.Query().Get("design")
		tag := r.URL.Query().Get("tag")
		web.ArticlesPageHandler(w, r, *s.Storage, tag, design)
	}))
	mux.Handle("/article", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		articleID := r.URL.Query().Get("id")
		web.GetArticleHandler(w, r, *s.Storage, articleID, s.auth)
	}))
	// Projects Routes
	mux.Handle("/projects", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		design := r.URL.Query().Get("design")
		tag := r.URL.Query().Get("tag")
		web.ProjectsPageHandler(w, r, *s.Storage, tag, design)
	}))
	mux.Handle("/project", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.URL.Query().Get("id")
		web.GetProjectHandler(w, r, *s.Storage, projectID, s.auth)
	}))
	// Reading List Routes
	mux.Handle("/reading-list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		design := r.URL.Query().Get("design")
		tag := r.URL.Query().Get("tag")
		web.ReadingListPageHandler(w, r, *s.Storage, tag, design)
	}))
	mux.Handle("/book", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		articleID := r.URL.Query().Get("id")
		web.GetReadingListBook(w, r, *s.Storage, articleID, s.auth)
	}))
	// Letter Routes
	mux.Handle("/letters", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		design := r.URL.Query().Get("design")
		tag := r.URL.Query().Get("tag")
		web.LettersPageHandler(w, r, *s.Storage, tag, design, s.auth)
	}))
	mux.Handle("/letter", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		letterID := r.URL.Query().Get("id")
		web.GetLetterHandler(w, r, *s.Storage, letterID, s.auth)
	}))
	// Wrap: recovery is outermost so it catches panics from all inner middleware.
	return recoveryMiddleware(
		securityHeadersMiddleware(
			s.corsMiddleware(s.maxBytesMiddleware(s.authContextMiddleware(mux))),
		),
	)
}

// authContextMiddleware resolves the request's auth state once and stores it in
// the context. Verifying the signed session cookie is not free, and templates
// need the answer as well as handlers, so doing it here avoids repeating the
// work for every component that asks.
func (s *Server) authContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := auth.WithAuthenticated(r.Context(), s.auth.IsAuthenticated(r))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				err := apperrors.PanicRecovered(fmt.Errorf("panic: %v", rec))
				web.HandleError(w, r, err, r.URL.Path, r.Method)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// maxBytesMiddleware limits the request body to 10MB on all routes to prevent
// memory exhaustion from oversized form submissions or request bodies.
func (s *Server) maxBytesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		next.ServeHTTP(w, r)
	})
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set(
			"Permissions-Policy",
			"camera=(), microphone=(), geolocation=()",
		)

		next.ServeHTTP(w, r)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", web.Site().URL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

// staticCacheMiddleware gives embedded assets a freshness lifetime. embed.FS
// reports a zero modtime, so without this they carry no cache headers at all
// and the browser refetches them on every navigation.
func staticCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", cacheControlFor(r.URL.Path))
		next.ServeHTTP(w, r)
	})
}

// cacheControlFor picks a lifetime from the asset's extension. Images and fonts
// change far less often than CSS and JS.
func cacheControlFor(urlPath string) string {
	switch strings.ToLower(path.Ext(urlPath)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".ico", ".woff", ".woff2":
		return longCacheControl
	default:
		return defaultCacheControl
	}
}

// HelloWorldHandler responds with a simple "Hello World" message ensuring server is running.
func (s *Server) HelloWorldHandler(w http.ResponseWriter, _ *http.Request) {
	resp := map[string]string{"message": "Hello World"}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonResp)
	if err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// HealthHandler responds with the current health status of the application.
func (s *Server) HealthHandler(w http.ResponseWriter, _ *http.Request) {
	result := s.Storage.Health()

	resp, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if !result.Healthy() {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
