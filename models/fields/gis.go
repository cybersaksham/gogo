package fields

import "fmt"

type GeometryKind string

const (
	Geometry           GeometryKind = "Geometry"
	Point              GeometryKind = "Point"
	LineString         GeometryKind = "LineString"
	Polygon            GeometryKind = "Polygon"
	MultiPoint         GeometryKind = "MultiPoint"
	MultiLineString    GeometryKind = "MultiLineString"
	MultiPolygon       GeometryKind = "MultiPolygon"
	GeometryCollection GeometryKind = "GeometryCollection"
)

// GISConfig configures geometry fields.
type GISConfig struct {
	SRID           int
	Geography      bool
	Dimensionality int
	SpatialIndex   bool
}

// GISField stores GIS metadata.
type GISField struct {
	*BaseField
	kind   GeometryKind
	config GISConfig
}

// NewGeometryField creates a generic geometry field.
func NewGeometryField(options Options, config GISConfig) *GISField {
	return newGISField(options, Geometry, config)
}

// NewPointField creates a point field.
func NewPointField(options Options, config GISConfig) *GISField {
	return newGISField(options, Point, config)
}

// NewLineStringField creates a line string field.
func NewLineStringField(options Options, config GISConfig) *GISField {
	return newGISField(options, LineString, config)
}

// NewPolygonField creates a polygon field.
func NewPolygonField(options Options, config GISConfig) *GISField {
	return newGISField(options, Polygon, config)
}

// NewMultiPointField creates a multi-point field.
func NewMultiPointField(options Options, config GISConfig) *GISField {
	return newGISField(options, MultiPoint, config)
}

// NewMultiLineStringField creates a multi-line string field.
func NewMultiLineStringField(options Options, config GISConfig) *GISField {
	return newGISField(options, MultiLineString, config)
}

// NewMultiPolygonField creates a multi-polygon field.
func NewMultiPolygonField(options Options, config GISConfig) *GISField {
	return newGISField(options, MultiPolygon, config)
}

// NewGeometryCollectionField creates a geometry collection field.
func NewGeometryCollectionField(options Options, config GISConfig) *GISField {
	return newGISField(options, GeometryCollection, config)
}

func newGISField(options Options, kind GeometryKind, config GISConfig) *GISField {
	if config.Dimensionality == 0 {
		config.Dimensionality = 2
	}
	return &GISField{BaseField: NewBaseField("gis", options, nil), kind: kind, config: config}
}

func (f *GISField) GeometryKind() GeometryKind {
	return f.kind
}

func (f *GISField) SRID() int {
	return f.config.SRID
}

func (f *GISField) Geography() bool {
	return f.config.Geography
}

func (f *GISField) Dimensionality() int {
	return f.config.Dimensionality
}

func (f *GISField) SpatialIndex() bool {
	return f.config.SpatialIndex
}

func (f *GISField) ColumnType(dialect string) string {
	if dialect != "postgres" && dialect != "postgresql" {
		return f.Kind()
	}
	storage := "geometry"
	if f.config.Geography {
		storage = "geography"
	}
	return fmt.Sprintf("%s(%s,%d)", storage, f.kind, f.config.SRID)
}

func (f *GISField) RequireDialect(dialect string) error {
	return requirePostgresDialect(dialect, f.Name())
}

func (f *GISField) ValidateMetadata() error {
	if f.config.SRID <= 0 {
		return fmt.Errorf("%w: SRID must be positive", ErrInvalidField)
	}
	if f.config.Dimensionality != 2 && f.config.Dimensionality != 3 && f.config.Dimensionality != 4 {
		return fmt.Errorf("%w: dimensionality must be 2, 3, or 4", ErrInvalidField)
	}
	return nil
}

func (f *GISField) Clone() Field {
	return &GISField{BaseField: f.BaseField.Clone().(*BaseField), kind: f.kind, config: f.config}
}

// RasterConfig configures raster fields.
type RasterConfig struct {
	SRID         int
	SpatialIndex bool
}

// RasterField stores raster metadata.
type RasterField struct {
	*BaseField
	config RasterConfig
}

// NewRasterField creates a raster field.
func NewRasterField(options Options, config RasterConfig) *RasterField {
	return &RasterField{BaseField: NewBaseField("raster", options, map[string]string{"postgres": "raster"}), config: config}
}

func (f *RasterField) SRID() int {
	return f.config.SRID
}

func (f *RasterField) SpatialIndex() bool {
	return f.config.SpatialIndex
}

func (f *RasterField) RequireDialect(dialect string) error {
	return requirePostgresDialect(dialect, f.Name())
}

func (f *RasterField) ValidateMetadata() error {
	if f.config.SRID <= 0 {
		return fmt.Errorf("%w: SRID must be positive", ErrInvalidField)
	}
	return nil
}

func (f *RasterField) Clone() Field {
	return &RasterField{BaseField: f.BaseField.Clone().(*BaseField), config: f.config}
}
