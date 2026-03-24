package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"time"
	"timterests/cmd/web/components"
	"timterests/internal/auth"
	"timterests/internal/model"
	"timterests/internal/storage"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// Article represents a blog article with metadata and content.
type Article struct {
	model.Document `yaml:",inline"`

	Date string `yaml:"date"`
}

// ArticlesPageHandler handles requests to the articles page,
// ensuring authentication and rendering the appropriate content.
func ArticlesPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, currentTag, design string) {
	var (
		component templ.Component
		tags      []string
	)

	articles, err := ListArticles(r.Context(), s, currentTag)
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
	articles, err := ListArticles(r.Context(), s, "all")
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

// ListArticles retrieves a list of articles from storage, optionally filtering by tag.
func ListArticles(ctx context.Context, s storage.Storage, tag string) ([]Article, error) {
	var articles []Article

	// Get all articles from the storage
	prefix := "articles/"

	articleFiles, err := s.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	for id, obj := range articleFiles {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		article, err := GetArticle(ctx, key, id, s)
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

// GetArticle retrieves a single article by its S3 key and ID.
func GetArticle(ctx context.Context, key string, id int, s storage.Storage) (*Article, error) {
	var article Article

	article.SetMeta(strconv.Itoa(id), key)

	err := s.GetPreparedFile(ctx, key, &article)
	if err != nil {
		return nil, fmt.Errorf("failed to get prepared file: %w", err)
	}

	return &article, nil
}

// GetLatestArticle retrieves the most recent article from storage.
func GetLatestArticle(ctx context.Context, s storage.Storage) (*Article, error) {
	articles, err := ListArticles(ctx, s, "all")
	if err != nil {
		return nil, err
	}

	if len(articles) > 0 {
		// Articles are sorted, first article is the latest
		latestArticle := articles[0]
		latestArticle.Body = storage.RemoveHTMLTags(latestArticle.Body)

		return &latestArticle, nil
	}

	return nil, errors.New("no articles found")
}

// ToCard converts an Article to a Card component for display in lists.
func (a Article) ToCard(i int) components.Card {
	return components.Card{
		Title:     a.Title,
		Subtitle:  a.Subtitle,
		Date:      a.Date,
		Body:      a.Body,
		ImagePath: "",
		Get:       "/article?id=" + a.ID,
		Tags:      a.Tags,
		Index:     i,
	}
}

// FormatDateForFilename converts a date string to a filename-safe format.
func FormatDateForFilename(dateStr string) string {
	// Parse the date string (assuming YYYY-MM-DD format)
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("FormatDateForFilename: failed to parse date %q: %v", dateStr, err)

		return dateStr
	}
	// Format as MM-DD-YYYY
	return t.Format("01-02-2006")
}
