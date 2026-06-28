package gis

import (
	"encoding/xml"
	"strings"
)

type GeoSitemapItem struct {
	Location string
	Geometry Geometry
}

func RenderGeoSitemap(items []GeoSitemapItem) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	builder.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:geo="http://www.google.com/geo/schemas/sitemap/1.0">`)
	for _, item := range items {
		builder.WriteString("<url><loc>")
		builder.WriteString(escapeXML(item.Location))
		builder.WriteString("</loc><geo:geo><geo:format>WKT</geo:format><geo:geom>")
		if item.Geometry != nil {
			builder.WriteString(escapeXML(item.Geometry.WKT()))
		}
		builder.WriteString("</geo:geom></geo:geo></url>")
	}
	builder.WriteString("</urlset>")
	return builder.String()
}

func escapeXML(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}
