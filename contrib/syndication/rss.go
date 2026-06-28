package syndication

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

func RenderRSS(_ *http.Request, feed Feed) string {
	var out bytes.Buffer
	_, _ = fmt.Fprint(&out, xml.Header, `<rss version="2.0"><channel>`)
	writeElement(&out, "title", feed.Title)
	writeElement(&out, "link", feed.Link)
	writeElement(&out, "description", feed.Description)
	writeElement(&out, "managingEditor", feed.Author)
	for _, category := range feed.Categories {
		writeElement(&out, "category", category)
	}
	writeElement(&out, "atom:link", feed.FeedURL)
	for _, item := range feed.Items {
		_, _ = fmt.Fprint(&out, "<item>")
		writeElement(&out, "title", item.Title)
		writeElement(&out, "description", item.Description)
		writeElement(&out, "link", item.Link)
		writeElement(&out, "author", item.Author)
		for _, category := range item.Categories {
			writeElement(&out, "category", category)
		}
		if item.PubDate != nil {
			writeElement(&out, "pubDate", item.PubDate.UTC().Format(time.RFC1123Z))
		}
		for _, enclosure := range item.Enclosures {
			_, _ = fmt.Fprintf(&out, `<enclosure url="%s" length="%d" type="%s"/>`, xmlEscape(enclosure.URL), enclosure.Length, xmlEscape(enclosure.Type))
		}
		_, _ = fmt.Fprint(&out, "</item>")
	}
	_, _ = fmt.Fprint(&out, "</channel></rss>")
	return out.String()
}

func writeElement(out *bytes.Buffer, name string, value string) {
	if value == "" {
		return
	}
	_, _ = fmt.Fprintf(out, "<%s>%s</%s>", name, xmlEscape(value), name)
}

func xmlEscape(value string) string {
	var out bytes.Buffer
	_ = xml.EscapeText(&out, []byte(value))
	return out.String()
}
