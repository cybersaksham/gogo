package sitemaps

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"
)

func IndexView(sections map[string]Sitemap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, _ = fmt.Fprint(w, xml.Header, `<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
		names := make([]string, 0, len(sections))
		for name := range sections {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			section := sections[name]
			pages := pageCount(len(section.Items()), section.Limit())
			for page := 1; page <= pages; page++ {
				_, _ = fmt.Fprintf(w, "<sitemap><loc>%s.xml?p=%d</loc></sitemap>", xmlEscape(name), page)
			}
		}
		_, _ = fmt.Fprint(w, "</sitemapindex>")
	})
}

func SectionView(section Sitemap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items := section.Items()
		limit := section.Limit()
		page, _ := strconv.Atoi(r.URL.Query().Get("p"))
		if page < 1 {
			page = 1
		}
		start := (page - 1) * limit
		end := start + limit
		if start > len(items) {
			start = len(items)
		}
		if end > len(items) {
			end = len(items)
		}
		items = items[start:end]
		if last := latestModified(items); !last.IsZero() {
			w.Header().Set("Last-Modified", last.Format(http.TimeFormat))
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, _ = fmt.Fprint(w, xml.Header, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">`)
		for _, item := range items {
			_, _ = fmt.Fprintf(w, "<url><loc>%s</loc>", xmlEscape(item.Location))
			if item.LastModified != nil {
				_, _ = fmt.Fprintf(w, "<lastmod>%s</lastmod>", item.LastModified.Format("2006-01-02"))
			}
			if item.ChangeFreq != "" {
				_, _ = fmt.Fprintf(w, "<changefreq>%s</changefreq>", xmlEscape(item.ChangeFreq))
			}
			if item.Priority > 0 {
				_, _ = fmt.Fprintf(w, "<priority>%.1f</priority>", item.Priority)
			}
			for _, alternate := range item.Alternates {
				_, _ = fmt.Fprintf(w, `<xhtml:link rel="alternate" hreflang="%s" href="%s"/>`, xmlEscape(alternate.Language), xmlEscape(alternate.Location))
			}
			_, _ = fmt.Fprint(w, "</url>")
		}
		_, _ = fmt.Fprint(w, "</urlset>")
	})
}

func pageCount(total int, limit int) int {
	if total == 0 {
		return 1
	}
	return (total + limit - 1) / limit
}

func latestModified(items []Item) time.Time {
	var latest time.Time
	for _, item := range items {
		if item.LastModified != nil && item.LastModified.After(latest) {
			latest = *item.LastModified
		}
	}
	return latest
}

func xmlEscape(value string) string {
	var out bytes.Buffer
	_ = xml.EscapeText(&out, []byte(value))
	return out.String()
}
