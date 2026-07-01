package migrations

import "github.com/cybersaksham/gogo/models"

// ProjectState stores historical migration state.
type ProjectState struct {
	Models map[string]ModelState
}

// ModelState stores historical model state without live Go types.
type ModelState struct {
	AppLabel    string            `json:"app_label"`
	Name        string            `json:"name"`
	TableName   string            `json:"table_name"`
	Fields      []FieldState      `json:"fields,omitempty"`
	Indexes     []IndexState      `json:"indexes,omitempty"`
	Constraints []ConstraintState `json:"constraints,omitempty"`
	Options     map[string]any    `json:"options,omitempty"`
}

// FieldState stores historical field state.
type FieldState struct {
	Name        string                  `json:"name"`
	Column      string                  `json:"column,omitempty"`
	Kind        string                  `json:"kind,omitempty"`
	ColumnTypes map[string]string       `json:"column_types,omitempty"`
	PrimaryKey  bool                    `json:"primary_key,omitempty"`
	Null        bool                    `json:"null,omitempty"`
	Unique      bool                    `json:"unique,omitempty"`
	DBIndex     bool                    `json:"db_index,omitempty"`
	DBDefault   *models.DatabaseDefault `json:"db_default,omitempty"`
	DBCollation string                  `json:"db_collation,omitempty"`
}

// IndexState stores historical index state.
type IndexState struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields,omitempty"`
}

// ConstraintState stores historical constraint state.
type ConstraintState struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Fields []string `json:"fields,omitempty"`
	Check  string   `json:"check,omitempty"`
}

// NewProjectState creates empty state.
func NewProjectState() ProjectState {
	return ProjectState{Models: make(map[string]ModelState)}
}

// Clone returns a deep state copy.
func (s ProjectState) Clone() ProjectState {
	copied := NewProjectState()
	for key, model := range s.Models {
		copied.Models[key] = model.clone()
	}
	return copied
}

// AddModel stores a model state.
func (s ProjectState) AddModel(model ModelState) {
	s.Models[modelKey(model.AppLabel, model.Name)] = model.clone()
}

// RemoveModel deletes a model state.
func (s ProjectState) RemoveModel(appLabel, modelName string) {
	delete(s.Models, modelKey(appLabel, modelName))
}

// AddField appends a field state.
func (s ProjectState) AddField(appLabel, modelName string, field FieldState) {
	key := modelKey(appLabel, modelName)
	model := s.Models[key].clone()
	model.Fields = append(model.Fields, cloneFieldState(field))
	s.Models[key] = model
}

// StateFromRegistry converts live model registry metadata into migration state.
func StateFromRegistry(registry *models.Registry) ProjectState {
	state := NewProjectState()
	if registry == nil {
		return state
	}
	for _, meta := range registry.Models() {
		if !meta.IsManaged() {
			continue
		}
		model := ModelState{
			AppLabel:    meta.AppLabel,
			Name:        meta.ModelName,
			TableName:   meta.TableName,
			Fields:      make([]FieldState, len(meta.Fields)),
			Indexes:     make([]IndexState, len(meta.Indexes)),
			Constraints: make([]ConstraintState, len(meta.Constraints)),
			Options: map[string]any{
				"verbose_name": meta.VerboseName,
				"ordering":     append([]string(nil), meta.Ordering...),
			},
		}
		for i, field := range meta.Fields {
			defaultValue, err := models.NormalizeDatabaseDefault(field.DBDefault)
			if err != nil {
				defaultValue = models.DatabaseDefault{}
			}
			model.Fields[i] = FieldState{
				Name:        field.Name,
				Column:      field.Column,
				Kind:        field.Kind,
				ColumnTypes: cloneStringMap(field.ColumnTypes),
				PrimaryKey:  field.PrimaryKey,
				Null:        field.Null,
				Unique:      field.Unique,
				DBIndex:     field.DBIndex,
				DBCollation: field.DBCollation,
			}
			if defaultValue.Kind != models.DefaultNone {
				model.Fields[i].DBDefault = &defaultValue
			}
		}
		for i, index := range meta.Indexes {
			model.Indexes[i] = IndexState{Name: index.Name, Fields: index.FieldNames()}
		}
		for i, constraint := range meta.Constraints {
			model.Constraints[i] = ConstraintState{Name: constraint.Name, Type: string(constraint.Type), Fields: constraint.FieldNames(), Check: constraint.Check}
		}
		state.AddModel(model)
	}
	return state
}

func (m ModelState) clone() ModelState {
	m.Fields = cloneFieldStates(m.Fields)
	m.Indexes = cloneIndexStates(m.Indexes)
	m.Constraints = cloneConstraintStates(m.Constraints)
	if m.Options != nil {
		options := make(map[string]any, len(m.Options))
		for key, value := range m.Options {
			options[key] = value
		}
		m.Options = options
	}
	return m
}

func cloneFieldStates(fields []FieldState) []FieldState {
	copied := make([]FieldState, len(fields))
	for i, field := range fields {
		copied[i] = cloneFieldState(field)
	}
	return copied
}

func cloneFieldState(field FieldState) FieldState {
	field.ColumnTypes = cloneStringMap(field.ColumnTypes)
	return field
}

func cloneIndexStates(indexes []IndexState) []IndexState {
	copied := make([]IndexState, len(indexes))
	for i, index := range indexes {
		copied[i] = IndexState{Name: index.Name, Fields: append([]string(nil), index.Fields...)}
	}
	return copied
}

func cloneConstraintStates(constraints []ConstraintState) []ConstraintState {
	copied := make([]ConstraintState, len(constraints))
	for i, constraint := range constraints {
		copied[i] = ConstraintState{
			Name:   constraint.Name,
			Type:   constraint.Type,
			Fields: append([]string(nil), constraint.Fields...),
			Check:  constraint.Check,
		}
	}
	return copied
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}

func modelKey(appLabel, modelName string) string {
	return appLabel + "." + modelName
}
