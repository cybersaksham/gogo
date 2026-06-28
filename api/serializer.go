package api

// Serializer validates input and renders output using ordered fields.
type Serializer struct {
	fields []SerializerField
}

// NewSerializer creates a serializer from fields.
func NewSerializer(fields ...SerializerField) *Serializer {
	copied := append([]SerializerField(nil), fields...)
	return &Serializer{fields: copied}
}

// Validate parses input into validated data and field errors.
func (s *Serializer) Validate(input map[string]any) (map[string]any, map[string][]string, bool) {
	validated := map[string]any{}
	fieldErrors := map[string][]string{}
	for _, field := range s.fields {
		if field.Options.ReadOnly {
			continue
		}
		value, exists := input[field.Name]
		if !exists || value == nil {
			if field.Options.Default != nil {
				validated[field.source()] = field.Options.Default
				continue
			}
			if field.Options.Required {
				fieldErrors[field.Name] = []string{"required"}
			}
			continue
		}
		if field.Kind == "nested" {
			nested, ok := value.(map[string]any)
			if !ok || field.Nested == nil {
				fieldErrors[field.Name] = []string{"invalid object"}
				continue
			}
			nestedValidated, nestedErrors, ok := field.Nested.Validate(nested)
			if !ok {
				for key, messages := range nestedErrors {
					fieldErrors[field.Name+"."+key] = messages
				}
				continue
			}
			validated[field.source()] = nestedValidated
			continue
		}
		parsed, errors := field.parse(value)
		if len(errors) > 0 {
			fieldErrors[field.Name] = errors
			continue
		}
		validated[field.source()] = parsed
	}
	return validated, fieldErrors, len(fieldErrors) == 0
}

// Render serializes an object map.
func (s *Serializer) Render(obj map[string]any) map[string]any {
	rendered := map[string]any{}
	for _, field := range s.fields {
		if field.Options.WriteOnly {
			continue
		}
		rendered[field.Name] = field.render(obj)
	}
	return rendered
}
