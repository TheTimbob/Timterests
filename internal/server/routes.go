package server

import (
	"encoding/json"
	"log"
	"net/http"

	"timterests/cmd/web"
	"timterests/internal/storage"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Favicon Route
	mux.Handle("/favicon.ico", http.FileServer(http.Dir(".")))

	// Serve static files from the "s3" directory
	mux.Handle("/s3/", http.StripPrefix("/s3/", http.FileServer(http.Dir("s3"))))

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

	mux.Handle("/writer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		docType := r.URL.Query().Get("document-type")
		if docType == "" {
			docType = "article" // default
		}

		web.WriterPageHandler(w, r, docType)
	}))

	mux.Handle("/write", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.WriteDocumentHandler(w, r, *s.storage)
	}))

	mux.Handle("/write/suggest", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.WriterSuggestionHandler(w, r)
	}))

	// Health check
	mux.HandleFunc("/health", s.healthHandler)

	// About Routes
	mux.Handle("/about", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.AboutHandler(w, r, *s.storage)
	}))

	// Login Routes
	mux.Handle("/login", http.HandlerFunc(web.LoginHandler))

	// Article Routes
	mux.Handle("/articles", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		design := r.URL.Query().Get("design")
		tag := r.URL.Query().Get("tag")
		web.ArticlesPageHandler(w, r, *s.storage, tag, design)
	}))
	mux.Handle("/article", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		articleID := r.URL.Query().Get("id")
		web.GetArticleHandler(w, r, *s.storage, articleID)
	}))
	mux.Handle("/articles/list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := web.ListArticles(*s.storage, "all"); err != nil {
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
		web.GetProjectHandler(w, r, *s.storage, projectID)
	}))
	mux.Handle("/projects/list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := web.ListProjects(*s.storage, "all"); err != nil {
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
		web.GetReadingListBook(w, r, *s.storage, articleID)
	}))
	mux.Handle("/reading-list/list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := web.ListBooks(*s.storage, "all"); err != nil {
			http.Error(w, "Failed to list articles", http.StatusInternalServerError)
			return
		}
	}))

	// Letter Routes
	mux.Handle("/letters", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		design := r.URL.Query().Get("design")
		tag := r.URL.Query().Get("tag")
		web.LettersPageHandler(w, r, *s.storage, tag, design)
	}))
	mux.Handle("/letter", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		letterID := r.URL.Query().Get("id")
		web.GetLetterHandler(w, r, *s.storage, letterID)
	}))
	mux.Handle("/letters/list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := web.ListLetters(*s.storage); err != nil {
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

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Hello World"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	//Marshal the health check response
	resp, err := json.Marshal(storage.Health(*s.storage))
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
