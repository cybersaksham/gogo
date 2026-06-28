package gis

type FieldSuggestion struct {
	Name      string
	FieldType string
}

func InspectModelFields(layer Layer) []FieldSuggestion {
	suggestions := []FieldSuggestion{{
		Name:      "geom",
		FieldType: geometryFieldType(layer.GeometryType),
	}}
	for _, field := range layer.Fields {
		suggestions = append(suggestions, FieldSuggestion{Name: field.Name, FieldType: scalarFieldType(field.Type)})
	}
	return suggestions
}

func geometryFieldType(kind string) string {
	switch kind {
	case "Point":
		return "PointField"
	case "LineString":
		return "LineStringField"
	case "Polygon":
		return "PolygonField"
	case "MultiPoint":
		return "MultiPointField"
	case "MultiLineString":
		return "MultiLineStringField"
	case "MultiPolygon":
		return "MultiPolygonField"
	default:
		return "GeometryField"
	}
}

func scalarFieldType(kind string) string {
	switch kind {
	case "Integer", "Integer64":
		return "IntegerField"
	case "Real", "Float":
		return "FloatField"
	case "Date":
		return "DateField"
	case "DateTime":
		return "DateTimeField"
	case "Boolean":
		return "BooleanField"
	default:
		return "CharField"
	}
}
