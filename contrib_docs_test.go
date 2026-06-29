package gogo

import (
	"os"
	"strings"
	"testing"
)

func TestContribDocsCoverInstallationMiddlewareAndDatabaseRequirements(t *testing.T) {
	const contribDocsPath = "docs/code/contrib.md"
	body, err := os.ReadFile(contribDocsPath)
	if err != nil {
		t.Fatalf("read %s: %v", contribDocsPath, err)
	}
	docs := string(body)
	for _, want := range []string{
		"InstalledApps",
		"gogo.contrib.sites",
		"gogo.contrib.redirects",
		"gogo.contrib.flatpages",
		"gogo.contrib.sitemaps",
		"gogo.contrib.syndication",
		"gogo.contrib.humanize",
		"gogo.contrib.admindocs",
		"gogo.contrib.postgres",
		"gogo.contrib.gis",
		"Middleware Order",
		"gogo.contrib.sites.Middleware",
		"gogo.messages.Middleware",
		"gogo.contrib.flatpages.Middleware",
		"gogo.contrib.redirects.Middleware",
		"PostgreSQL Extension Requirements",
		"pg_trgm",
		"postgis",
		"GIS Database Requirements",
		"AllowUnsafeTargets",
		"go run manage.go check",
	} {
		if !strings.Contains(docs, want) {
			t.Fatalf("%s does not document %s", contribDocsPath, want)
		}
	}
}
