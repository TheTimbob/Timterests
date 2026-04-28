package web

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"

	"timterests/internal/service"
	"timterests/internal/storage"
)

type urlSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// RobotsHandler serves robots.txt.
func RobotsHandler(w http.ResponseWriter, _ *http.Request) {
	baseURL := siteURL()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	body := fmt.Sprintf(
		"User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /writer\nDisallow: /login\nDisallow: /download\n\nSitemap: %s/sitemap.xml\n",
		baseURL,
	)

	_, err := w.Write([]byte(body))
	if err != nil {
		log.Printf("RobotsHandler: failed to write response: %v", err)
	}
}

// SitemapHandler generates and serves sitemap.xml.
func SitemapHandler(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	baseURL := siteURL()
	now := time.Now().Format("2006-01-02")

	staticPages := []sitemapURL{
		{Loc: baseURL + "/", Priority: "1.0", ChangeFreq: "weekly", LastMod: now},
		{Loc: baseURL + "/articles", Priority: "0.9", ChangeFreq: "weekly", LastMod: now},
		{Loc: baseURL + "/projects", Priority: "0.9", ChangeFreq: "monthly", LastMod: now},
		{Loc: baseURL + "/reading-list", Priority: "0.7", ChangeFreq: "monthly", LastMod: now},
		{Loc: baseURL + "/about", Priority: "0.8", ChangeFreq: "monthly", LastMod: now},
	}

	articles, err := service.ListArticles(r.Context(), s, "all")
	if err != nil {
		log.Printf("SitemapHandler: failed to list articles: %v", err)
	}

	projects, err := service.ListProjects(r.Context(), s, "all")
	if err != nil {
		log.Printf("SitemapHandler: failed to list projects: %v", err)
	}

	books, err := service.ListBooks(r.Context(), s, "all")
	if err != nil {
		log.Printf("SitemapHandler: failed to list books: %v", err)
	}

	urls := make([]sitemapURL, 0, len(staticPages)+len(articles)+len(projects)+len(books))
	urls = append(urls, staticPages...)

	for _, a := range articles {
		urls = append(urls, sitemapURL{
			Loc:        baseURL + "/article?id=" + a.ID,
			LastMod:    a.Date,
			ChangeFreq: "yearly",
			Priority:   "0.8",
		})
	}

	for _, p := range projects {
		urls = append(urls, sitemapURL{
			Loc:        baseURL + "/project?id=" + p.ID,
			ChangeFreq: "monthly",
			Priority:   "0.7",
		})
	}

	for _, b := range books {
		urls = append(urls, sitemapURL{
			Loc:        baseURL + "/book?id=" + b.ID,
			ChangeFreq: "yearly",
			Priority:   "0.5",
		})
	}

	sitemap := urlSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	_, err = w.Write([]byte(xml.Header))
	if err != nil {
		log.Printf("SitemapHandler: failed to write XML header: %v", err)

		return
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")

	err = enc.Encode(sitemap)
	if err != nil {
		log.Printf("SitemapHandler: failed to encode sitemap: %v", err)
	}
}
