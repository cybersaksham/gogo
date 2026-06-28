package gis

import "sort"

type LayerMapping struct {
	Layer       Layer
	Model       string
	Mapping     map[string]string
	Transform   func(Geometry) (Geometry, error)
	Encoding    string
	Unique      []string
	Transaction bool
	Progress    func(done int, total int)
}

type ImportPlan struct {
	Model       string
	Layer       string
	Fields      []MappedField
	Encoding    string
	Unique      []string
	Transaction bool
}

type MappedField struct {
	Source string
	Target string
}

func (l LayerMapping) Plan() ImportPlan {
	fields := make([]MappedField, 0, len(l.Mapping))
	sources := make([]string, 0, len(l.Mapping))
	for source := range l.Mapping {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	for _, source := range sources {
		fields = append(fields, MappedField{Source: source, Target: l.Mapping[source]})
	}
	if l.Progress != nil {
		l.Progress(l.Layer.FeatureCount, l.Layer.FeatureCount)
	}
	return ImportPlan{
		Model:       l.Model,
		Layer:       l.Layer.Name,
		Fields:      fields,
		Encoding:    l.Encoding,
		Unique:      append([]string(nil), l.Unique...),
		Transaction: l.Transaction,
	}
}
