package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"slices"
	"strconv"
	"timterests/internal/models"
	"timterests/internal/storage"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func ArticlesPageHandler(w http.ResponseWriter, r *http.Request, storageInstance models.Storage, tag string) {
    var component templ.Component
    var tags []string

    articles, err := ListArticles(storageInstance, tag)
	if err != nil {
		message := "Failed to fetch articles"
		http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusInternalServerError)
		return
	}

    for i := range articles {
		articles[i].Body = storage.RemoveHTMLTags(articles[i].Body)
        tags = GetTags(articles[i], tags)
    }

    if tag == "all" {
	    component = ArticlesPage(articles, tags)
    } else {
        component = ArticlesList(articles)
    }
    
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in ArticlesPosts: %e", err)
	}
}

func GetArticleHandler(w http.ResponseWriter, r *http.Request, storageInstance models.Storage, articleID string) {

	articles, err := ListArticles(storageInstance, "all")
	if err != nil {
		http.Error(w, "Failed to fetch articles", http.StatusInternalServerError)
		return
	}

	for _, article := range articles {
		if article.ID == articleID {
			component := ArticlePage(article)
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Fatalf("Error rendering in GetArticleByIDHandler: %e", err)
			}
		}
	}

}

func ListArticles(storageInstance models.Storage, tag string) ([]models.Document, error) {
	var articles []models.Document
	var article models.Document

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

		fileName := path.Base(key)
		localFilePath := path.Join("s3", fileName)

		article, err = storage.ReadFile(key, localFilePath, storageInstance)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
			return nil, err
		}

		article.ID = strconv.Itoa(id)
        if slices.Contains(article.Tags, tag) || tag == "all" {
		    articles = append(articles, article)
        }
	}

	return articles, nil
}

func GetTags(article models.Document, tags []string) []string {

    for _, tag := range article.Tags {
        if !slices.Contains(tags, tag) {
            tags = append(tags, tag)
        }
    }

    return tags
}