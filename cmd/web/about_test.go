package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"timterests/cmd/web"

	"github.com/PuerkitoBio/goquery"
)

func TestAboutHandler(t *testing.T) {
	s := testSetup(t, context.Background())

	t.Run("renders about page", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/about", nil)
		rec := httptest.NewRecorder()

		web.AboutHandler(rec, req, *s)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		doc, err := goquery.NewDocumentFromReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if doc.Find("title").Length() == 0 {
			t.Error("expected title element to be rendered")
		}
	})
}
