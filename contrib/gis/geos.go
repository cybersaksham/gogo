package gis

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Geometry interface {
	WKT() string
	WKB() []byte
	HEXEWKB() string
	GeoJSON() string
	SRIDValue() int
	Bounds() Bounds
}

type Bounds struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

func (b Bounds) Intersects(other Bounds) bool {
	return b.MinX <= other.MaxX && b.MaxX >= other.MinX && b.MinY <= other.MaxY && b.MaxY >= other.MinY
}

func (b Bounds) Contains(other Bounds) bool {
	return b.MinX <= other.MinX && b.MaxX >= other.MaxX && b.MinY <= other.MinY && b.MaxY >= other.MaxY
}

func (b Bounds) Equals(other Bounds) bool {
	return b.MinX == other.MinX && b.MinY == other.MinY && b.MaxX == other.MaxX && b.MaxY == other.MaxY
}

type Point struct {
	X    float64
	Y    float64
	SRID int
}

func (p Point) WKT() string {
	return withSRID(p.SRID, p.wktBody())
}

func (p Point) WKB() []byte {
	return []byte(p.WKT())
}

func (p Point) HEXEWKB() string {
	return hex.EncodeToString(p.WKB())
}

func (p Point) GeoJSON() string {
	return fmt.Sprintf(`{"type":"Point","coordinates":[%s,%s]}`, formatFloat(p.X), formatFloat(p.Y))
}

func (p Point) SRIDValue() int {
	return p.SRID
}

func (p Point) Bounds() Bounds {
	return Bounds{MinX: p.X, MinY: p.Y, MaxX: p.X, MaxY: p.Y}
}

func (p Point) wktBody() string {
	return fmt.Sprintf("POINT(%s %s)", formatFloat(p.X), formatFloat(p.Y))
}

type LineString struct {
	Points []Point
	SRID   int
}

func (l LineString) WKT() string {
	return withSRID(l.SRID, l.wktBody())
}

func (l LineString) WKB() []byte {
	return []byte(l.WKT())
}

func (l LineString) HEXEWKB() string {
	return hex.EncodeToString(l.WKB())
}

func (l LineString) GeoJSON() string {
	parts := make([]string, len(l.Points))
	for i, point := range l.Points {
		parts[i] = fmt.Sprintf("[%s,%s]", formatFloat(point.X), formatFloat(point.Y))
	}
	return fmt.Sprintf(`{"type":"LineString","coordinates":[%s]}`, strings.Join(parts, ","))
}

func (l LineString) SRIDValue() int {
	return l.SRID
}

func (l LineString) Bounds() Bounds {
	return boundsForPoints(l.Points)
}

func (l LineString) wktBody() string {
	parts := make([]string, len(l.Points))
	for i, point := range l.Points {
		parts[i] = fmt.Sprintf("%s %s", formatFloat(point.X), formatFloat(point.Y))
	}
	return "LINESTRING(" + strings.Join(parts, ",") + ")"
}

type Polygon struct {
	Rings [][]Point
	SRID  int
}

func (p Polygon) WKT() string {
	return withSRID(p.SRID, p.wktBody())
}

func (p Polygon) WKB() []byte {
	return []byte(p.WKT())
}

func (p Polygon) HEXEWKB() string {
	return hex.EncodeToString(p.WKB())
}

func (p Polygon) GeoJSON() string {
	rings := make([]string, len(p.Rings))
	for i, ring := range p.Rings {
		points := make([]string, len(ring))
		for j, point := range ring {
			points[j] = fmt.Sprintf("[%s,%s]", formatFloat(point.X), formatFloat(point.Y))
		}
		rings[i] = "[" + strings.Join(points, ",") + "]"
	}
	return fmt.Sprintf(`{"type":"Polygon","coordinates":[%s]}`, strings.Join(rings, ","))
}

func (p Polygon) SRIDValue() int {
	return p.SRID
}

func (p Polygon) Bounds() Bounds {
	var points []Point
	for _, ring := range p.Rings {
		points = append(points, ring...)
	}
	return boundsForPoints(points)
}

func (p Polygon) wktBody() string {
	rings := make([]string, len(p.Rings))
	for i, ring := range p.Rings {
		points := make([]string, len(ring))
		for j, point := range ring {
			points[j] = fmt.Sprintf("%s %s", formatFloat(point.X), formatFloat(point.Y))
		}
		rings[i] = "(" + strings.Join(points, ",") + ")"
	}
	return "POLYGON(" + strings.Join(rings, ",") + ")"
}

type GeometryCollection struct {
	Geometries []Geometry
	SRID       int
}

func (g GeometryCollection) WKT() string {
	return withSRID(g.SRID, g.wktBody())
}

func (g GeometryCollection) WKB() []byte {
	return []byte(g.WKT())
}

func (g GeometryCollection) HEXEWKB() string {
	return hex.EncodeToString(g.WKB())
}

func (g GeometryCollection) GeoJSON() string {
	parts := make([]string, len(g.Geometries))
	for i, geometry := range g.Geometries {
		parts[i] = geometry.GeoJSON()
	}
	return fmt.Sprintf(`{"type":"GeometryCollection","geometries":[%s]}`, strings.Join(parts, ","))
}

func (g GeometryCollection) SRIDValue() int {
	return g.SRID
}

func (g GeometryCollection) Bounds() Bounds {
	if len(g.Geometries) == 0 {
		return Bounds{}
	}
	bounds := g.Geometries[0].Bounds()
	for _, geometry := range g.Geometries[1:] {
		bounds = expandBounds(bounds, geometry.Bounds())
	}
	return bounds
}

func (g GeometryCollection) wktBody() string {
	parts := make([]string, len(g.Geometries))
	for i, geometry := range g.Geometries {
		parts[i] = wktBody(geometry)
	}
	return "GEOMETRYCOLLECTION(" + strings.Join(parts, ",") + ")"
}

type PreparedGeometry struct {
	geometry Geometry
}

func Prepared(geometry Geometry) PreparedGeometry {
	return PreparedGeometry{geometry: geometry}
}

func (p PreparedGeometry) Contains(other Geometry) bool {
	return p.geometry.Bounds().Contains(other.Bounds())
}

func (p PreparedGeometry) CoveredBy(other Geometry) bool {
	return Prepared(other).Covers(p.geometry)
}

func (p PreparedGeometry) Covers(other Geometry) bool {
	return p.Contains(other) || p.Equals(other)
}

func (p PreparedGeometry) Crosses(other Geometry) bool {
	return p.Intersects(other) && !p.Contains(other) && !Prepared(other).Contains(p.geometry)
}

func (p PreparedGeometry) Disjoint(other Geometry) bool {
	return !p.Intersects(other)
}

func (p PreparedGeometry) Equals(other Geometry) bool {
	return p.geometry.SRIDValue() == other.SRIDValue() && p.geometry.WKT() == other.WKT()
}

func (p PreparedGeometry) Intersects(other Geometry) bool {
	return p.geometry.Bounds().Intersects(other.Bounds())
}

func (p PreparedGeometry) Overlaps(other Geometry) bool {
	return p.Intersects(other) && !p.Contains(other) && !Prepared(other).Contains(p.geometry)
}

func (p PreparedGeometry) Relate(other Geometry, pattern string) bool {
	if pattern == "T********" {
		return p.Intersects(other)
	}
	if pattern == "F********" {
		return p.Disjoint(other)
	}
	return p.Intersects(other)
}

func (p PreparedGeometry) Touches(other Geometry) bool {
	bounds := p.geometry.Bounds()
	otherBounds := other.Bounds()
	xTouches := bounds.MaxX == otherBounds.MinX || bounds.MinX == otherBounds.MaxX
	yOverlaps := bounds.MinY <= otherBounds.MaxY && bounds.MaxY >= otherBounds.MinY
	yTouches := bounds.MaxY == otherBounds.MinY || bounds.MinY == otherBounds.MaxY
	xOverlaps := bounds.MinX <= otherBounds.MaxX && bounds.MaxX >= otherBounds.MinX
	return (xTouches && yOverlaps) || (yTouches && xOverlaps)
}

func (p PreparedGeometry) Within(other Geometry) bool {
	return Prepared(other).Contains(p.geometry)
}

func GeometryFromWKT(input string) (Geometry, error) {
	srid := 0
	wkt := strings.TrimSpace(input)
	if strings.HasPrefix(strings.ToUpper(wkt), "SRID=") {
		parts := strings.SplitN(wkt, ";", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid EWKT: %s", input)
		}
		parsedSRID, err := strconv.Atoi(strings.TrimPrefix(strings.ToUpper(parts[0]), "SRID="))
		if err != nil {
			return nil, fmt.Errorf("invalid SRID: %w", err)
		}
		srid = parsedSRID
		wkt = strings.TrimSpace(parts[1])
	}
	upper := strings.ToUpper(wkt)
	if strings.HasPrefix(upper, "POINT(") && strings.HasSuffix(wkt, ")") {
		body := strings.TrimSuffix(strings.TrimPrefix(wkt, wkt[:6]), ")")
		coordinates := strings.Fields(body)
		if len(coordinates) != 2 {
			return nil, fmt.Errorf("invalid POINT coordinates: %s", input)
		}
		x, err := strconv.ParseFloat(coordinates[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid POINT x coordinate: %w", err)
		}
		y, err := strconv.ParseFloat(coordinates[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid POINT y coordinate: %w", err)
		}
		return Point{X: x, Y: y, SRID: srid}, nil
	}
	return nil, fmt.Errorf("unsupported WKT geometry: %s", input)
}

func withSRID(srid int, body string) string {
	if srid <= 0 {
		return body
	}
	return fmt.Sprintf("SRID=%d;%s", srid, body)
}

func wktBody(geometry Geometry) string {
	switch value := geometry.(type) {
	case Point:
		return value.wktBody()
	case LineString:
		return value.wktBody()
	case Polygon:
		return value.wktBody()
	case GeometryCollection:
		return value.wktBody()
	default:
		return geometry.WKT()
	}
}

func boundsForPoints(points []Point) Bounds {
	if len(points) == 0 {
		return Bounds{}
	}
	bounds := Bounds{MinX: math.Inf(1), MinY: math.Inf(1), MaxX: math.Inf(-1), MaxY: math.Inf(-1)}
	for _, point := range points {
		if point.X < bounds.MinX {
			bounds.MinX = point.X
		}
		if point.Y < bounds.MinY {
			bounds.MinY = point.Y
		}
		if point.X > bounds.MaxX {
			bounds.MaxX = point.X
		}
		if point.Y > bounds.MaxY {
			bounds.MaxY = point.Y
		}
	}
	return bounds
}

func expandBounds(left Bounds, right Bounds) Bounds {
	return Bounds{
		MinX: math.Min(left.MinX, right.MinX),
		MinY: math.Min(left.MinY, right.MinY),
		MaxX: math.Max(left.MaxX, right.MaxX),
		MaxY: math.Max(left.MaxY, right.MaxY),
	}
}
