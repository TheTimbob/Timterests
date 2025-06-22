package web

import (
	"log"
	"net/http"
	"timterests/internal/storage"
	"timterests/internal/types"
)

type Article struct {
	types.Document `yaml:",inline"`
	Date           string `yaml:"date"`
}

func ArticlesPageHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, currentTag, design string) {
	component, err := ListPageHandler[*Article](storageInstance, currentTag, "/articles", design)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in ArticlesPosts: %e", err)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in ArticlesPosts: %e", err)
	}
}

func GetArticleHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, articleID string) {
	component, err := ItemPageHandler[*Article](storageInstance, articleID, "/articles", true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in GetArticleByIDHandler: %e", err)
	}
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in GetArticleByIDHandler: %e", err)
	}
}
