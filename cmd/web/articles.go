package web

import (
	"context"
	"html"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"timterests/internal/models"
	"timterests/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"gopkg.in/yaml.v2"
)

func ArticlesListHandler(w http.ResponseWriter, r *http.Request, storageInstance *storage.Storage) {
	articles, err := ListArticles(storageInstance)
	if err != nil {
		http.Error(w, "Failed to fetch articles", http.StatusInternalServerError)
		return
	}

	component := ArticlesList(articles)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in ArticlesPosts: %e", err)
	}
}

func GetArticleHandler(w http.ResponseWriter, r *http.Request, storageInstance *storage.Storage, articleID string) {

	articles, err := ListArticles(storageInstance)
	if err != nil {
		http.Error(w, "Failed to fetch articles", http.StatusInternalServerError)
		return
	}

	for _, article := range articles {
		if article.ID == articleID {
			article.Body = ConvertTextToParagraphs(article.Body)
			component := ArticlePage(article)
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Fatalf("Error rendering in GetArticleByIDHandler: %e", err)
			}
		}
	}

}

func ListArticles(storageInstance *storage.Storage) ([]models.Article, error) {
	var articles []models.Article
	var article models.Article

	// Get all articles from the storage
	prefix := "articles/"
	articleFiles, err := storageInstance.ListObjects(context.Background(), prefix)
	if err != nil {
		return nil, err
	}

	for id, obj := range articleFiles {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		fileName := path.Base(key)
		localFilePath := path.Join("tmp", fileName)

		// Download the file
		err := storageInstance.DownloadFile(context.Background(), key, localFilePath)
		if err != nil {
			log.Println("Failed to download file: ", err)
			return nil, err
		}

		// Open the downloaded file
		file, err := os.Open(localFilePath)
		if err != nil {
			log.Println("Failed to open file: ", err)
			return nil, err
		}
		defer file.Close()

		// Decode the yaml file into an article object
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(&article); err != nil {
			log.Println("Failed to decode file: ", err)
			return nil, err
		}

		article.ID = strconv.Itoa(id)
		articles = append(articles, article)
	}

	return articles, nil
}

// Converts raw text into HTML paragraphs
func ConvertTextToParagraphs(text string) string {
	paragraphs := strings.Split(text, "\n\n") // Split by double newline for paragraphs
	var htmlContent string

	for _, paragraph := range paragraphs {
		// Escape any special HTML characters to prevent injection
		htmlContent += "<p class='content-text'>" + html.EscapeString(paragraph) + "</p>"
	}

	return htmlContent
}
