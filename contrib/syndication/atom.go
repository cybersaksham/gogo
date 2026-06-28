package syndication

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

func RenderAtom(_ *http.Request, feed Feed) string {
	var out bytes.Buffer
	_, _ = fmt.Fprint(&out, xml.Header, `<feed xmlns="http://www.w3.org/2005/Atom">`)
	writeElement(&out, "title", feed.Title)
	if feed.Link != "" {
		_, _ = fmt.Fprintf(&out, `<link href="%s"/>`, xmlEscape(feed.Link))
	}
	if feed.FeedURL != "" {
		_, _ = fmt.Fprintf(&out, `<link rel="self" href="%s"/>`, xmlEscape(feed.FeedURL))
	}
	writeElement(&out, "subtitle", feed.Description)
	writeElement(&out, "id", feed.Link)
	for _, category := range feed.Categories {
		_, _ = fmt.Fprintf(&out, `<category term="%s"/>`, xmlEscape(category))
	}
	for _, item := range feed.Items {
		_, _ = fmt.Fprint(&out, "<entry>")
		writeElement(&out, "title", item.Title)
		writeElement(&out, "id", item.Link)
		if item.Link != "" {
			_, _ = fmt.Fprintf(&out, `<link href="%s"/>`, xmlEscape(item.Link))
		}
		if item.Updated != nil {
			writeElement(&out, "updated", item.Updated.UTC().Format(time.RFC3339))
		} else if item.PubDate != nil {
			writeElement(&out, "updated", item.PubDate.UTC().Format(time.RFC3339))
		}
		writeElement(&out, "summary", item.Description)
		if item.Author != "" {
			_, _ = fmt.Fprintf(&out, "<author><name>%s</name></author>", xmlEscape(item.Author))
		}
		for _, category := range item.Categories {
			_, _ = fmt.Fprintf(&out, `<category term="%s"/>`, xmlEscape(category))
		}
		for _, enclosure := range item.Enclosures {
			_, _ = fmt.Fprintf(&out, `<link rel="enclosure" href="%s" length="%d" type="%s"/>`, xmlEscape(enclosure.URL), enclosure.Length, xmlEscape(enclosure.Type))
		}
		_, _ = fmt.Fprint(&out, "</entry>")
	}
	_, _ = fmt.Fprint(&out, "</feed>")
	return out.String()
}
