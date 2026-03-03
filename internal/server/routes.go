// Package server provides HTTP server and routing configuration.
package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"timterests/cmd/web"
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
		web.HomeHandler(w, r, *s.storage)
	}))
	mux.Handle("/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.HomeHandler(w, r, *s.storage)
	}))
	mux.Handle("/web", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.HomeHandler(w, r, *s.storage)
	}))
	mux.Handle("/web/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.HomeHandler(w, r, *s.storage)
	}))

	mux.Handle("/admin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.AdminPageHandler(w, r, s.auth)
	}))

	mux.Handle("/writer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			docType, key string
			typeID       int
		)

		// Handle POST request - parse form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)

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
				http.Error(w, "Invalid type ID: expected integer, got '"+typeIDString+"'", http.StatusBadRequest)

				return
			}
		}

		key = r.FormValue("document-key")

		web.WriterPageHandler(w, r, *s.storage, docType, key, typeID, s.auth)
	}))

	mux.Handle("/write", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.WriteDocumentHandler(w, r, *s.storage, s.auth)
	}))

	mux.Handle("/write/suggest", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.WriterSuggestionHandler(w, r, *s.storage, s.auth)
	}))

	mux.Handle("/download", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		documentTitle := r.URL.Query().Get("title")
		web.DownloadDocumentHandler(w, r, documentTitle, s.auth)
	}))

	mux.Handle("/download/new", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.DownloadNewDocumentHandler(w, r, s.auth)
	}))

	// Health check
	mux.HandleFunc("/health", s.healthHandler)

	// About Routes
	mux.Handle("/about", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.AboutHandler(w, r, *s.storage)
	}))

	// Login Routes
	mux.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.LoginHandler(w, r, s.auth)
	}))

	// Article Routes
	mux.Handle("/articles", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		design := r.URL.Query().Get("design")
		tag := r.URL.Query().Get("tag")
		web.ArticlesPageHandler(w, r, *s.storage, tag, design)
	}))
	mux.Handle("/article", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		articleID := r.URL.Query().Get("id")
		web.GetArticleHandler(w, r, *s.storage, articleID, s.auth)
	}))
	mux.Handle("/articles/list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := web.ListArticles(r.Context(), *s.storage, "all")
		if err != nil {
			http.Error(w, "Failed to list articles", http.StatusInternalServerError)

			return
		}
	}))

	// Projects Routes
	mux.Handle("/projects", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		design := r.URL.Query().Get("design")
		tag := r.URL.Query().Get("tag")
		web.ProjectsPageHandler(w, r, *s.storage, tag, design)
	}))
	mux.Handle("/project", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.URL.Query().Get("id")
		web.GetProjectHandler(w, r, *s.storage, projectID, s.auth)
	}))
	mux.Handle("/projects/list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := web.ListProjects(r.Context(), *s.storage, "all")
		if err != nil {
			http.Error(w, "Failed to list articles", http.StatusInternalServerError)

			return
		}
	}))

	// Reading List Routes
	mux.Handle("/reading-list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		design := r.URL.Query().Get("design")
		tag := r.URL.Query().Get("tag")
		web.ReadingListPageHandler(w, r, *s.storage, tag, design)
	}))
	mux.Handle("/book", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		articleID := r.URL.Query().Get("id")
		web.GetReadingListBook(w, r, *s.storage, articleID, s.auth)
	}))
	mux.Handle("/reading-list/list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := web.ListBooks(r.Context(), *s.storage, "all")
		if err != nil {
			http.Error(w, "Failed to list articles", http.StatusInternalServerError)

			return
		}
	}))

	// Letter Routes
	mux.Handle("/letters", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		design := r.URL.Query().Get("design")
		tag := r.URL.Query().Get("tag")
		web.LettersPageHandler(w, r, *s.storage, tag, design, s.auth)
	}))
	mux.Handle("/letter", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		letterID := r.URL.Query().Get("id")
		web.GetLetterHandler(w, r, *s.storage, letterID, s.auth)
	}))
	mux.Handle("/letters/list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := web.ListLetters(r.Context(), *s.storage, "all")
		if err != nil {
			http.Error(w, "Failed to list letters.", http.StatusInternalServerError)

			return
		}
	}))

	// Wrap the mux with CORS middleware
	return s.corsMiddleware(mux)
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

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	// Marshal the health check response
	resp, err := json.Marshal(s.storage.Health())
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
