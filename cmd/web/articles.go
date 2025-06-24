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

func (a Article) GetID() string       { return a.ID }
func (a Article) GetBody() string     { return a.Body }
func (a Article) GetTitle() string    { return a.Title }
func (a Article) GetSubtitle() string { return a.Subtitle }
func (a Article) GetTags() []string   { return a.Tags }

func ArticlesPageHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, currentTag, design string) {
	component, err := GetListPageComponent[Article](storageInstance, currentTag, "article", design)
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

	page := r.Header.Get("HX-Request") != "true"
	component, err := GetItemComponent[Article](storageInstance, articleID, "article", page)
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
