package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"timterests/internal/models"
	"timterests/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func ArticlesListHandler(w http.ResponseWriter, r *http.Request, storageInstance models.Storage) {
	articles, err := ListArticles(storageInstance)
	if err != nil {
		message := "Failed to fetch articles"
		http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusInternalServerError)
		return
	}

	component := ArticlesList(articles)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in ArticlesPosts: %e", err)
	}
}

func GetArticleHandler(w http.ResponseWriter, r *http.Request, storageInstance models.Storage, articleID string) {

	articles, err := ListArticles(storageInstance)
	if err != nil {
		http.Error(w, "Failed to fetch articles", http.StatusInternalServerError)
		return
	}

	for _, article := range articles {
		if article.ID == articleID {
			article.Body = storage.ConvertTextToParagraphs(article.Body)
			component := ArticlePage(article)
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Fatalf("Error rendering in GetArticleByIDHandler: %e", err)
			}
		}
	}

}

func ListArticles(storageInstance models.Storage) ([]models.Document, error) {
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
			log.Printf("Failed to read file: %v", err)
			return nil, err
		}

		article.ID = strconv.Itoa(id)
		articles = append(articles, article)
	}

	return articles, nil
}
