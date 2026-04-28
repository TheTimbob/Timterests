package web

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"timterests/internal/service"
	"timterests/internal/storage"
)

type urlSet struct {
	XMLName xml.Name  `xml:"urlset"`
	XMLNS   string    `xml:"xmlns,attr"`
	URLs    []siteURL `xml:"url"`
}

type siteURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func siteBaseURL() string {
	if u := os.Getenv("SITE_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}

	return "https://timterests.com"
}

// RobotsHandler serves robots.txt.
func RobotsHandler(w http.ResponseWriter, _ *http.Request) {
	baseURL := siteBaseURL()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	fmt.Fprintf(w, "User-agent: *\n")
	fmt.Fprintf(w, "Allow: /\n")
	fmt.Fprintf(w, "Disallow: /admin\n")
	fmt.Fprintf(w, "Disallow: /writer\n")
	fmt.Fprintf(w, "Disallow: /login\n")
	fmt.Fprintf(w, "Disallow: /download\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Sitemap: %s/sitemap.xml\n", baseURL)
}

// SitemapHandler generates and serves sitemap.xml.
func SitemapHandler(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	baseURL := siteBaseURL()
	now := time.Now().Format("2006-01-02")

	urls := []siteURL{
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

	for _, a := range articles {
		urls = append(urls, siteURL{
			Loc:        baseURL + "/article?id=" + a.ID,
			LastMod:    a.Date,
			ChangeFreq: "yearly",
			Priority:   "0.8",
		})
	}

	projects, err := service.ListProjects(r.Context(), s, "all")
	if err != nil {
		log.Printf("SitemapHandler: failed to list projects: %v", err)
	}

	for _, p := range projects {
		urls = append(urls, siteURL{
			Loc:        baseURL + "/project?id=" + p.ID,
			ChangeFreq: "monthly",
			Priority:   "0.7",
		})
	}

	books, err := service.ListBooks(r.Context(), s, "all")
	if err != nil {
		log.Printf("SitemapHandler: failed to list books: %v", err)
	}

	for _, b := range books {
		urls = append(urls, siteURL{
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

	if _, err := w.Write([]byte(xml.Header)); err != nil {
		log.Printf("SitemapHandler: failed to write XML header: %v", err)

		return
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")

	if err := enc.Encode(sitemap); err != nil {
		log.Printf("SitemapHandler: failed to encode sitemap: %v", err)
	}
}
