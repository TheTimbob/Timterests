package web

import (
	"encoding/xml"
	"log"
	"net/http"
	"strings"
	"time"

	"timterests/internal/service"
	"timterests/internal/storage"
)

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	Items       []rssItem `xml:"item"`
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate,omitempty"`
	GUID        string `xml:"guid"`
}

// RSSHandler generates and serves an RSS 2.0 feed of articles.
func RSSHandler(
	w http.ResponseWriter,
	r *http.Request,
	s storage.Storage,
) {
	baseURL := strings.TrimRight(Site().URL, "/")

	articles, err := service.ListArticles(r.Context(), s, "all")
	if err != nil {
		log.Printf("RSSHandler: failed to list articles: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)

		return
	}

	items := make([]rssItem, 0, len(articles))

	for _, a := range articles {
		desc := a.Subtitle
		if desc == "" {
			desc = a.Preview
		}

		var pubDate string

		t, err := time.Parse("2006-01-02", a.Date)
		if err == nil {
			pubDate = t.Format(time.RFC1123Z)
		}

		items = append(items, rssItem{
			Title:       a.Title,
			Link:        baseURL + "/article?id=" + a.ID,
			Description: desc,
			PubDate:     pubDate,
			GUID:        baseURL + "/article?id=" + a.ID,
		})
	}

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:       Site().Name,
			Link:        baseURL,
			Description: Site().Description,
			Language:    "en",
			Items:       items,
		},
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")

	_, err = w.Write([]byte(xml.Header))
	if err != nil {
		log.Printf("RSSHandler: failed to write XML header: %v", err)

		return
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")

	err = enc.Encode(feed)
	if err != nil {
		log.Printf("RSSHandler: failed to encode RSS feed: %v", err)
	}
}
