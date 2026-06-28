package sitemaps

import "time"

type Alternate struct {
	Language string
	Location string
}

type Item struct {
	Location     string
	LastModified *time.Time
	ChangeFreq   string
	Priority     float64
	Alternates   []Alternate
}

type Sitemap interface {
	Items() []Item
	Limit() int
	ProtocolValue() string
}

type StaticSitemap struct {
	Protocol   string
	LimitValue int
	ItemsValue []Item
}

func (s StaticSitemap) Items() []Item {
	return append([]Item(nil), s.ItemsValue...)
}

func (s StaticSitemap) Limit() int {
	if s.LimitValue <= 0 {
		return 50000
	}
	return s.LimitValue
}

func (s StaticSitemap) ProtocolValue() string {
	if s.Protocol == "" {
		return "http"
	}
	return s.Protocol
}
