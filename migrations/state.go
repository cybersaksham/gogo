package migrations

import "github.com/cybersaksham/gogo/models"

// ProjectState stores historical migration state.
type ProjectState struct {
	Models map[string]ModelState
}

// ModelState stores historical model state without live Go types.
type ModelState struct {
	AppLabel    string
	Name        string
	TableName   string
	Fields      []FieldState
	Indexes     []IndexState
	Constraints []ConstraintState
	Options     map[string]any
}

// FieldState stores historical field state.
type FieldState struct {
	Name       string
	Column     string
	Kind       string
	PrimaryKey bool
	Null       bool
	Unique     bool
}

// IndexState stores historical index state.
type IndexState struct {
	Name   string
	Fields []string
}

// ConstraintState stores historical constraint state.
type ConstraintState struct {
	Name   string
	Type   string
	Fields []string
	Check  string
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
	model.Fields = append(model.Fields, field)
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
			model.Fields[i] = FieldState{
				Name:       field.Name,
				Column:     field.Column,
				PrimaryKey: field.PrimaryKey,
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
	m.Fields = append([]FieldState(nil), m.Fields...)
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

func modelKey(appLabel, modelName string) string {
	return appLabel + "." + modelName
}
