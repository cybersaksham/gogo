package syndication

import "time"

type Feed struct {
	Title       string
	Link        string
	Description string
	Author      string
	Categories  []string
	FeedURL     string
	Items       []Item
}

type Item struct {
	Title       string
	Description string
	Link        string
	PubDate     *time.Time
	Updated     *time.Time
	Author      string
	Categories  []string
	Enclosures  []Enclosure
}

type Enclosure struct {
	URL    string
	Length int64
	Type   string
}

func (f Feed) ForObject(name string, items []Item) Feed {
	clone := f.Clone()
	if name != "" {
		clone.Title = clone.Title + ": " + name
	}
	clone.Items = append([]Item(nil), items...)
	return clone
}

func (f Feed) Clone() Feed {
	clone := f
	clone.Categories = append([]string(nil), f.Categories...)
	clone.Items = make([]Item, len(f.Items))
	for i, item := range f.Items {
		clone.Items[i] = item.Clone()
	}
	return clone
}

func (i Item) Clone() Item {
	clone := i
	clone.Categories = append([]string(nil), i.Categories...)
	clone.Enclosures = append([]Enclosure(nil), i.Enclosures...)
	if i.PubDate != nil {
		pub := *i.PubDate
		clone.PubDate = &pub
	}
	if i.Updated != nil {
		updated := *i.Updated
		clone.Updated = &updated
	}
	return clone
}
