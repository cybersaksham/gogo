package fields

import (
	"errors"
	"testing"
)

func TestGISGeometryConstructorsExposeMetadata(t *testing.T) {
	config := GISConfig{SRID: 4326, Geography: true, Dimensionality: 2, SpatialIndex: true}
	tests := []struct {
		name  string
		field *GISField
		kind  GeometryKind
	}{
		{"geometry", NewGeometryField(Options{Name: "geom"}, config), Geometry},
		{"point", NewPointField(Options{Name: "geom"}, config), Point},
		{"line", NewLineStringField(Options{Name: "geom"}, config), LineString},
		{"polygon", NewPolygonField(Options{Name: "geom"}, config), Polygon},
		{"multipoint", NewMultiPointField(Options{Name: "geom"}, config), MultiPoint},
		{"multiline", NewMultiLineStringField(Options{Name: "geom"}, config), MultiLineString},
		{"multipolygon", NewMultiPolygonField(Options{Name: "geom"}, config), MultiPolygon},
		{"collection", NewGeometryCollectionField(Options{Name: "geom"}, config), GeometryCollection},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.field.GeometryKind() != test.kind {
				t.Fatalf("GeometryKind() = %s, want %s", test.field.GeometryKind(), test.kind)
			}
			if test.field.SRID() != 4326 || !test.field.Geography() || test.field.Dimensionality() != 2 || !test.field.SpatialIndex() {
				t.Fatalf("metadata = %#v", test.field)
			}
			if err := test.field.RequireDialect("sqlite"); !errors.Is(err, ErrInvalidField) {
				t.Fatalf("RequireDialect(sqlite) error = %v, want ErrInvalidField", err)
			}
		})
	}
}

func TestGISFieldRejectsInvalidMetadata(t *testing.T) {
	field := NewPointField(Options{Name: "point"}, GISConfig{SRID: 0, Dimensionality: 5})
	if err := field.ValidateMetadata(); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("ValidateMetadata() error = %v, want ErrInvalidField", err)
	}
}

func TestRasterFieldMetadata(t *testing.T) {
	field := NewRasterField(Options{Name: "raster"}, RasterConfig{SRID: 3857, SpatialIndex: true})

	if field.SRID() != 3857 || !field.SpatialIndex() {
		t.Fatalf("raster metadata = %#v", field)
	}
	if got := field.ColumnType("postgres"); got != "raster" {
		t.Fatalf("ColumnType(postgres) = %q, want raster", got)
	}
}
