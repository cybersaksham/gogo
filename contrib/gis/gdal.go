package gis

type DataSource struct {
	Path   string
	Layers []Layer
}

func (d DataSource) LayerByName(name string) (Layer, bool) {
	for _, layer := range d.Layers {
		if layer.Name == name {
			return layer, true
		}
	}
	return Layer{}, false
}

type Layer struct {
	Name         string
	GeometryType string
	Fields       []Field
	FeatureCount int
	SpatialRef   string
}

type Field struct {
	Name string
	Type string
}

type CoordinateTransform struct {
	SourceSRID int
	TargetSRID int
	Transform  func(Geometry) (Geometry, error)
}

func (c CoordinateTransform) Apply(geometry Geometry) (Geometry, error) {
	if c.Transform == nil {
		return geometry, nil
	}
	return c.Transform(geometry)
}
