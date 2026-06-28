package sitemaps

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSitemapIndexSectionPaginationLastModifiedAlternatesAndEscaping(t *testing.T) {
	last := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	site := StaticSitemap{Protocol: "https", LimitValue: 1, ItemsValue: []Item{
		{Location: "/a?x=<tag>", LastModified: &last, ChangeFreq: "daily", Priority: 0.8, Alternates: []Alternate{{Language: "fr", Location: "/fr/a"}}},
		{Location: "/b", LastModified: &last},
	}}
	index := httptest.NewRecorder()
	IndexView(map[string]Sitemap{"main": site}).ServeHTTP(index, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if !strings.Contains(index.Body.String(), "main.xml?p=1") || !strings.Contains(index.Body.String(), "main.xml?p=2") {
		t.Fatalf("index XML = %s", index.Body.String())
	}
	section := httptest.NewRecorder()
	SectionView(site).ServeHTTP(section, httptest.NewRequest(http.MethodGet, "/main.xml?p=1", nil))
	body := section.Body.String()
	if !strings.Contains(body, "&lt;tag&gt;") || !strings.Contains(body, `hreflang="fr"`) || section.Header().Get("Last-Modified") == "" {
		t.Fatalf("section XML = %s headers=%#v", body, section.Header())
	}
}
