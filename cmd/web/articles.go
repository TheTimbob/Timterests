package web

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"timterests/internal/auth"
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
		message := "Failed to fetch articles"
		http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusInternalServerError)

		return
	}

	for i := range articles {
		articles[i].Body = storage.RemoveHTMLTags(articles[i].Body)
		v := reflect.ValueOf(articles[i])
		tags = storage.GetTags(v, tags)
	}

	if r.Header.Get("Hx-Request") == "true" {
		component = ArticlesList(articles, design)
	} else {
		component = ArticlesListPage(articles, tags, design)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in ArticlesPosts: %e", err)
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
			var component templ.Component

			authenticated := a.IsAuthenticated(r)

			if r.Header.Get("Hx-Request") == "true" {
				component = ArticleDisplay(article, authenticated)
			} else {
				component = ArticlePage(article, authenticated)
			}

			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Printf("Error rendering in GetArticleByIDHandler: %e", err)
			}
		}
	}
}

// FormatDateForFilename converts a date string to a filename-safe format.
func FormatDateForFilename(dateStr string) string {
	return service.FormatArticleDateForFilename(dateStr)
}
