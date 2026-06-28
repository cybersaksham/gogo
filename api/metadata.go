package api

// MetadataOptions configures endpoint metadata generation.
type MetadataOptions struct {
	Serializer     *Serializer
	FilterSet      FilterSet
	Pagination     any
	Authentication []string
	Permissions    []string
	Throttles      []string
}

// APIMetadata describes routes and policies for browsable API, OPTIONS, and schema generation.
type APIMetadata struct {
	Version        string
	BrowsableAPI   bool
	Routes         []RouteMetadata
	Actions        []ActionMetadata
	Serializer     SerializerMetadata
	Forms          map[string][]SerializerFieldMetadata
	Filters        FilterMetadata
	Pagination     string
	Authentication []string
	Permissions    []string
	Throttles      []string
}

// RouteMetadata describes one API route.
type RouteMetadata struct {
	Name    string
	Pattern string
	Methods []string
	Action  string
	Detail  bool
}

// ActionMetadata describes one viewset action.
type ActionMetadata struct {
	Name    string
	Methods []string
	Detail  bool
}

// SerializerMetadata describes serializer fields.
type SerializerMetadata struct {
	Fields []SerializerFieldMetadata
}

// SerializerFieldMetadata describes one serializer field.
type SerializerFieldMetadata struct {
	Name      string
	Kind      string
	Source    string
	Required  bool
	ReadOnly  bool
	WriteOnly bool
	Label     string
	HelpText  string
	Choices   []string
}

// FilterMetadata describes filter controls.
type FilterMetadata struct {
	Exact    []string
	Lookups  map[string][]string
	Search   []string
	Ordering []string
}

// BuildMetadata returns API endpoint metadata from router and component options.
func BuildMetadata(request *Request, router *Router, options MetadataOptions) APIMetadata {
	metadata := APIMetadata{
		BrowsableAPI:   true,
		Forms:          map[string][]SerializerFieldMetadata{},
		Authentication: append([]string(nil), options.Authentication...),
		Permissions:    append([]string(nil), options.Permissions...),
		Throttles:      append([]string(nil), options.Throttles...),
	}
	if request != nil {
		metadata.Version = request.Version()
	}
	if router != nil {
		for _, route := range router.Routes() {
			metadata.Routes = append(metadata.Routes, RouteMetadata{
				Name:    route.Name,
				Pattern: route.Pattern,
				Methods: append([]string(nil), route.Methods...),
				Action:  route.Action,
				Detail:  route.Detail,
			})
			if route.Action != "" {
				metadata.Actions = append(metadata.Actions, ActionMetadata{
					Name:    route.Action,
					Methods: append([]string(nil), route.Methods...),
					Detail:  route.Detail,
				})
			}
		}
	}
	metadata.Serializer = SerializerMetadataFor(options.Serializer)
	fields := metadata.Serializer.Fields
	metadata.Forms["create"] = cloneFieldMetadata(fields)
	metadata.Forms["update"] = cloneFieldMetadata(fields)
	metadata.Forms["partial_update"] = cloneFieldMetadata(fields)
	metadata.Filters = FilterMetadata{
		Exact:    append([]string(nil), options.FilterSet.ExactFields...),
		Lookups:  cloneLookupMetadata(options.FilterSet.LookupFields),
		Search:   append([]string(nil), options.FilterSet.SearchFields...),
		Ordering: append([]string(nil), options.FilterSet.OrderingFields...),
	}
	metadata.Pagination = paginationMetadataName(options.Pagination)
	return metadata
}

// SerializerMetadataFor returns metadata for serializer fields.
func SerializerMetadataFor(serializer *Serializer) SerializerMetadata {
	if serializer == nil {
		return SerializerMetadata{}
	}
	fields := make([]SerializerFieldMetadata, 0, len(serializer.fields))
	for _, field := range serializer.fields {
		fields = append(fields, SerializerFieldMetadata{
			Name:      field.Name,
			Kind:      field.Kind,
			Source:    field.source(),
			Required:  field.Options.Required,
			ReadOnly:  field.Options.ReadOnly,
			WriteOnly: field.Options.WriteOnly,
			Label:     field.Options.Label,
			HelpText:  field.Options.HelpText,
			Choices:   append([]string(nil), field.Choices...),
		})
	}
	return SerializerMetadata{Fields: fields}
}

func cloneLookupMetadata(values map[string][]string) map[string][]string {
	if values == nil {
		return nil
	}
	copied := make(map[string][]string, len(values))
	for key, value := range values {
		copied[key] = append([]string(nil), value...)
	}
	return copied
}

func cloneFieldMetadata(values []SerializerFieldMetadata) []SerializerFieldMetadata {
	copied := make([]SerializerFieldMetadata, len(values))
	copy(copied, values)
	for index := range copied {
		copied[index].Choices = append([]string(nil), values[index].Choices...)
	}
	return copied
}

func paginationMetadataName(value any) string {
	switch value.(type) {
	case PageNumberPagination:
		return "page_number"
	case LimitOffsetPagination:
		return "limit_offset"
	case CursorPagination:
		return "cursor"
	default:
		return ""
	}
}
