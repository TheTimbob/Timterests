package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"timterests/cmd/web"
	apperrors "timterests/internal/errors"
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
	mux.Handle("/assets/", fileServer)

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

	mux.Handle("/admin/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.AdminUsersPageHandler(w, r, s.auth)
	}))
	mux.Handle("/admin/users/create", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.CreateUserHandler(w, r, s.auth)
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

	mux.Handle("/write/suggest", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.WriterSuggestionHandler(w, r, *s.Storage, s.auth)
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

	// Health check
	mux.HandleFunc("/health", s.HealthHandler)

	// About Routes
	mux.Handle("/about", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.AboutHandler(w, r, *s.Storage)
	}))

	// Login Routes
	mux.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.LoginHandler(w, r, s.auth)
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
	return recoveryMiddleware(s.corsMiddleware(s.maxBytesMiddleware(mux)))
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

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
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
