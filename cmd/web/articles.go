package web

import (
	"net/http"
	"reflect"

	apperrors "timterests/internal/errors"

	"timterests/internal/auth"
	"timterests/internal/model"
	"timterests/internal/service"
	"timterests/internal/storage"

	"github.com/a-h/templ"
)

func ArticlesPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, currentTag, design string) {
	var (
		component templ.Component
		tags      []string
	)

	articles, err := service.ListArticles(r.Context(), s, currentTag)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "ArticlesPageHandler", "listArticles")

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
		HandleError(w, r, apperrors.RenderFailed(err), "ArticlesPageHandler", "render")
	}
}

func GetArticleHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, articleID string, a *auth.Auth) {
	articles, err := service.ListArticles(r.Context(), s, "all")
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "GetArticleHandler", "listArticles")

		return
	}

	for _, article := range articles {
		if article.ID == articleID {
			body, err := s.GetDocumentBody(r.Context(), article.S3Key)
			if err != nil {
				HandleError(w, r, apperrors.NotFound(err), "GetArticleHandler", "getBody")

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
				HandleError(w, r, apperrors.RenderFailed(err), "GetArticleHandler", "render")

				return
			}

			return
		}
	}

	HandleError(w, r, apperrors.NotFound(nil), "GetArticleHandler", "findArticle")
}

func FormatDateForFilename(dateStr string) string {
	return service.FormatArticleDateForFilename(dateStr)
}
