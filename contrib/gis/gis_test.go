package gis

import (
	"errors"
	"strings"
	"testing"
)

func TestGeometryWrappersSerializeAndPreparePredicates(t *testing.T) {
	point := Point{X: 77.2, Y: 28.6, SRID: 4326}
	if got := point.WKT(); got != "SRID=4326;POINT(77.2 28.6)" {
		t.Fatalf("Point.WKT() = %q", got)
	}
	if got := point.HEXEWKB(); got == "" {
		t.Fatal("Point.HEXEWKB() is empty")
	}
	if got := point.GeoJSON(); !strings.Contains(got, `"type":"Point"`) || !strings.Contains(got, `"coordinates":[77.2,28.6]`) {
		t.Fatalf("Point.GeoJSON() = %q", got)
	}

	parsed, err := GeometryFromWKT("SRID=4326;POINT(77.2 28.6)")
	if err != nil {
		t.Fatalf("GeometryFromWKT() error = %v", err)
	}
	if !Prepared(parsed).Equals(point) {
		t.Fatal("parsed point should equal original point")
	}

	polygon := Polygon{
		Rings: [][]Point{{
			{X: 77, Y: 28, SRID: 4326},
			{X: 78, Y: 28, SRID: 4326},
			{X: 78, Y: 29, SRID: 4326},
			{X: 77, Y: 29, SRID: 4326},
			{X: 77, Y: 28, SRID: 4326},
		}},
		SRID: 4326,
	}
	if !Prepared(polygon).Contains(point) || !Prepared(point).Within(polygon) {
		t.Fatal("prepared polygon should contain point and point should be within polygon")
	}

	collection := GeometryCollection{Geometries: []Geometry{point, LineString{Points: []Point{{X: 0, Y: 0}, {X: 1, Y: 1}}}}, SRID: 4326}
	if got := collection.WKT(); !strings.Contains(got, "GEOMETRYCOLLECTION(") || !strings.Contains(got, "LINESTRING(0 0,1 1)") {
		t.Fatalf("GeometryCollection.WKT() = %q", got)
	}
}

func TestSpatialFunctionsLookupsMeasurementsAndUnsupportedDialects(t *testing.T) {
	sql, err := Area("geom").SQL("postgres")
	if err != nil || sql != `ST_Area("geom")` {
		t.Fatalf("Area SQL = %q, %v", sql, err)
	}
	sql, err = Distance("geom", "$1").SQL("postgres")
	if err != nil || sql != `ST_Distance("geom", $1)` {
		t.Fatalf("Distance SQL = %q, %v", sql, err)
	}
	sql, err = Transform("geom", 3857).SQL("postgres")
	if err != nil || sql != `ST_Transform("geom", 3857)` {
		t.Fatalf("Transform SQL = %q, %v", sql, err)
	}
	if _, err := AsGeoJSON("geom").SQL("sqlite"); !errors.Is(err, ErrUnsupportedDialect) {
		t.Fatalf("AsGeoJSON sqlite error = %v, want ErrUnsupportedDialect", err)
	}

	sql, err = Contains("geom", "$1").SQL("postgres")
	if err != nil || sql != `ST_Contains("geom", $1)` {
		t.Fatalf("Contains SQL = %q, %v", sql, err)
	}
	sql, err = DistanceLTE("geom", "$1", DistanceMeasure{Meters: 1000}).SQL("postgres")
	if err != nil || sql != `ST_DWithin("geom", $1, 1000)` {
		t.Fatalf("DistanceLTE SQL = %q, %v", sql, err)
	}
	sql, err = Relate("geom", "$1", "T********").SQL("postgres")
	if err != nil || sql != `ST_Relate("geom", $1, 'T********')` {
		t.Fatalf("Relate SQL = %q, %v", sql, err)
	}
	if _, err := Intersects("geom", "$1").SQL("mysql"); !errors.Is(err, ErrUnsupportedDialect) {
		t.Fatalf("Intersects mysql error = %v, want ErrUnsupportedDialect", err)
	}

	if got := D(1500).Kilometers(); got != 1.5 {
		t.Fatalf("D(1500).Kilometers() = %v", got)
	}
	if got := A(1000000).SquareKilometers(); got != 1 {
		t.Fatalf("A(1000000).SquareKilometers() = %v", got)
	}
}

func TestGDALLayerMappingInspectAndGeoSitemap(t *testing.T) {
	source := DataSource{
		Path: "/data/cities.gpkg",
		Layers: []Layer{{
			Name:         "cities",
			GeometryType: "Point",
			SpatialRef:   "EPSG:4326",
			FeatureCount: 2,
			Fields: []Field{
				{Name: "name", Type: "String"},
				{Name: "population", Type: "Integer"},
			},
		}},
	}
	layer, ok := source.LayerByName("cities")
	if !ok || layer.FeatureCount != 2 || layer.SpatialRef != "EPSG:4326" {
		t.Fatalf("LayerByName() = %#v, %v", layer, ok)
	}

	suggestions := InspectModelFields(layer)
	if len(suggestions) != 3 || suggestions[0].Name != "geom" || suggestions[0].FieldType != "PointField" || suggestions[2].FieldType != "IntegerField" {
		t.Fatalf("InspectModelFields() = %#v", suggestions)
	}

	progressCalls := 0
	mapping := LayerMapping{
		Layer:       layer,
		Model:       "City",
		Mapping:     map[string]string{"name": "Name", "population": "Population"},
		Encoding:    "utf-8",
		Unique:      []string{"name"},
		Transaction: true,
		Progress: func(done int, total int) {
			progressCalls++
			if done != 2 || total != 2 {
				t.Fatalf("progress = %d/%d", done, total)
			}
		},
	}
	plan := mapping.Plan()
	if len(plan.Fields) != 2 || plan.Fields[0].Source != "name" || !plan.Transaction || plan.Encoding != "utf-8" {
		t.Fatalf("LayerMapping.Plan() = %#v", plan)
	}
	if progressCalls != 1 {
		t.Fatalf("progress calls = %d", progressCalls)
	}

	xml := RenderGeoSitemap([]GeoSitemapItem{{
		Location: "https://example.com/cities/delhi",
		Geometry: Point{X: 77.2, Y: 28.6, SRID: 4326},
	}})
	if !strings.Contains(xml, `<geo:format>WKT</geo:format>`) || !strings.Contains(xml, `SRID=4326;POINT(77.2 28.6)`) {
		t.Fatalf("RenderGeoSitemap() = %q", xml)
	}
}
