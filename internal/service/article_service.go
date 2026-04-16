// Package service provides business logic for data retrieval and manipulation,
// decoupled from HTTP handlers.
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"
	"time"
	"timterests/internal/model"
	"timterests/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// ListArticles retrieves all articles from storage, optionally filtering by tag.
// Pass tag="" or tag="all" to retrieve all articles.
func ListArticles(ctx context.Context, s storage.Storage, tag string) ([]model.Article, error) {
	var articles []model.Article

	prefix := "articles/"

	articleFiles, err := s.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects with prefix %q: %w", prefix, err)
	}

	docIdx := 0

	for _, obj := range articleFiles {
		key := aws.ToString(obj.Key)

		if key == prefix || !strings.HasSuffix(key, ".yaml") {
			continue
		}

		if !s.FileExists(ctx, strings.TrimSuffix(key, ".yaml")+".md") {
			log.Printf("ListArticles: skipping %s — no paired .md body file", key)

			continue
		}

		article, err := GetArticle(ctx, s, key, docIdx)
		if err != nil {
			return nil, err
		}

		docIdx++

		if tag == "all" || tag == "" || slices.Contains(article.Tags, tag) {
			articles = append(articles, *article)
		}
	}

	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date > articles[j].Date
	})

	return articles, nil
}

// GetArticle retrieves a single article by its storage key and numeric ID.
func GetArticle(ctx context.Context, s storage.Storage, key string, id int) (*model.Article, error) {
	return getDoc[model.Article](ctx, s, key, id)
}

// GetLatestArticle retrieves the most recently dated article from storage.
// Articles are sorted by date descending; the first one is the latest.
func GetLatestArticle(ctx context.Context, s storage.Storage) (*model.Article, error) {
	articles, err := ListArticles(ctx, s, "all")
	if err != nil {
		return nil, err
	}

	if len(articles) == 0 {
		return nil, errors.New("no articles found")
	}

	// Articles are sorted by date desc — first element is latest.
	latestArticle := articles[0]

	return &latestArticle, nil
}

// FormatArticleDateForFilename converts a date string (YYYY-MM-DD) to MM-DD-YYYY
// for use in filenames.
func FormatArticleDateForFilename(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("FormatArticleDateForFilename: failed to parse date %q: %v", dateStr, err)

		return dateStr
	}

	return t.Format("01-02-2006")
}
