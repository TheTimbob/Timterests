package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"timterests/internal/storage"
	"timterests/internal/types"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type Article struct {
	types.Document `yaml:",inline"`
	Date           string `yaml:"date"`
}

func ArticlesPageHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, currentTag, design string) {
	var component templ.Component
	var tags []string

	articles, err := ListArticles(storageInstance, currentTag)
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

	if r.Header.Get("HX-Request") == "true" {
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

func GetArticleHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, articleID string) {
	articles, err := ListArticles(storageInstance, "all")
	if err != nil {
		http.Error(w, "Failed to fetch articles", http.StatusInternalServerError)
		return
	}

	for _, article := range articles {
		if article.ID == articleID {
			var component templ.Component
			if r.Header.Get("HX-Request") == "true" {
				component = ArticleDisplay(article)
			} else {
				component = ArticlePage(article)
			}
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Printf("Error rendering in GetArticleByIDHandler: %e", err)
			}
		}
	}

}

func ListArticles(storageInstance storage.Storage, tag string) ([]Article, error) {
	var articles []Article

	// Get all articles from the storage
	prefix := "articles/"
	articleFiles, err := storage.ListObjects(context.Background(), storageInstance, prefix)
	if err != nil {
		return nil, err
	}

	for id, obj := range articleFiles {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		article, err := GetArticle(key, id, storageInstance)
		if err != nil {
			return nil, err
		}

		if slices.Contains(article.Tags, tag) || tag == "all" || tag == "" {
			articles = append(articles, *article)
		}
	}

	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date > articles[j].Date
	})

	return articles, nil
}

func GetArticle(key string, id int, storageInstance storage.Storage) (*Article, error) {
	var article Article
	article.ID = strconv.Itoa(id)
	err := storage.GetPreparedFile(key, &article, storageInstance)
	if err != nil {
		return nil, err
	}

	return &article, nil
}

func GetLatestArticle(storageInstance storage.Storage) (*Article, error) {

	articles, err := ListArticles(storageInstance, "all")
	if err != nil {
		return nil, err
	}

	if len(articles) > 0 {
		// Articles are sorted, first article is the latest
		latestArticle := articles[0]
		latestArticle.Body = storage.RemoveHTMLTags(latestArticle.Body)
		return &latestArticle, nil
	}

	return nil, nil
}
