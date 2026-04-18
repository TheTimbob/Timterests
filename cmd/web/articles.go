package web

import (
	"log"
	"net/http"
	"reflect"
	"timterests/internal/auth"
	"timterests/internal/model"
	"timterests/internal/service"
	"timterests/internal/storage"

	"github.com/a-h/templ"
)

// ArticlesPageHandler handles requests to the articles page,
// ensuring authentication and rendering the appropriate content.
func ArticlesPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, currentTag, design string) {
	var (
		component templ.Component
		tags      []string
	)

	articles, err := service.ListArticles(r.Context(), s, currentTag)
	if err != nil {
		log.Printf("ArticlesPageHandler: failed to fetch articles: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	for i := range articles {
		v := reflect.ValueOf(articles[i])
		tags = storage.GetTags(v, tags)
	}

	if r.Header.Get("Hx-Request") == "true" {
		component = ArticlesList(articles, design)
	} else {
		component = ArticlesListPage(articles, tags, design)
	}

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		log.Printf("ArticlesPageHandler: failed to render: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetArticleHandler retrieves and renders a specific article by its ID.
func GetArticleHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, articleID string, a *auth.Auth) {
	articles, err := service.ListArticles(r.Context(), s, "all")
	if err != nil {
		http.Error(w, "Failed to fetch articles", http.StatusInternalServerError)

		return
	}

	for _, article := range articles {
		if article.ID == articleID {
			body, err := s.GetDocumentBody(r.Context(), article.S3Key)
			if err != nil {
				log.Printf("GetArticleHandler: failed to load body for %s: %v", article.S3Key, err)
				http.Error(w, "Not Found", http.StatusNotFound)

				return
			}

			dc := model.DisplayContent{
				ID:    article.ID,
				S3Key: article.S3Key,
				Body:  body,
			}

			var component templ.Component

			authenticated := a.IsAuthenticated(r)

			if r.Header.Get("Hx-Request") == "true" {
				component = ArticleDisplay(dc, authenticated)
			} else {
				component = ArticlePage(dc, authenticated)
			}

			err = renderHTML(w, r, http.StatusOK, component)
			if err != nil {
				log.Printf("GetArticleHandler: failed to render: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			return
		}
	}

	http.Error(w, "Not Found", http.StatusNotFound)
}

// FormatDateForFilename converts a date string to a filename-safe format.
func FormatDateForFilename(dateStr string) string {
	return service.FormatArticleDateForFilename(dateStr)
}
