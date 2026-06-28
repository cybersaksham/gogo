package syndication

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRSSAtomItemsEnclosuresObjectSpecificAndEscaping(t *testing.T) {
	pub := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	feed := Feed{
		Title:       "News <Today>",
		Link:        "https://example.com/news",
		Description: "Latest",
		Author:      "Admin",
		Categories:  []string{"updates"},
		FeedURL:     "https://example.com/feed.xml",
		Items: []Item{{
			Title:       "Item <One>",
			Description: "Body",
			Link:        "https://example.com/item",
			PubDate:     &pub,
			Updated:     &pub,
			Author:      "Author",
			Categories:  []string{"cat"},
			Enclosures:  []Enclosure{{URL: "https://example.com/a.mp3", Length: 12, Type: "audio/mpeg"}},
		}},
	}
	rss := RenderRSS(httptest.NewRequest("GET", "/feed", nil), feed)
	if !strings.Contains(rss, "<rss version=\"2.0\"") || !strings.Contains(rss, "News &lt;Today&gt;") || !strings.Contains(rss, "<enclosure") {
		t.Fatalf("rss = %s", rss)
	}
	atom := RenderAtom(httptest.NewRequest("GET", "/feed", nil), feed)
	if !strings.Contains(atom, `<feed xmlns="http://www.w3.org/2005/Atom">`) || !strings.Contains(atom, "<entry>") {
		t.Fatalf("atom = %s", atom)
	}
	objectFeed := feed.ForObject("section", []Item{{Title: "Object", Link: "/object"}})
	if objectFeed.Title != "News <Today>: section" || len(objectFeed.Items) != 1 {
		t.Fatalf("object feed = %#v", objectFeed)
	}
}
