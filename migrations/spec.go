package migrations

import "encoding/json"

// OperationSpec stores canonical operation content for generated migrations.
type OperationSpec struct {
	Type               string           `json:"type"`
	AppLabel           string           `json:"app_label,omitempty"`
	ModelName          string           `json:"model_name,omitempty"`
	Model              *ModelState      `json:"model,omitempty"`
	Field              *FieldState      `json:"field,omitempty"`
	OldField           *FieldState      `json:"old_field,omitempty"`
	NewField           *FieldState      `json:"new_field,omitempty"`
	OldName            string           `json:"old_name,omitempty"`
	NewName            string           `json:"new_name,omitempty"`
	OldTable           string           `json:"old_table,omitempty"`
	NewTable           string           `json:"new_table,omitempty"`
	Comment            string           `json:"comment,omitempty"`
	Options            map[string]any   `json:"options,omitempty"`
	Managers           []string         `json:"managers,omitempty"`
	FieldName          string           `json:"field_name,omitempty"`
	UniqueTogether     [][]string       `json:"unique_together,omitempty"`
	IndexTogether      [][]string       `json:"index_together,omitempty"`
	Index              *IndexState      `json:"index,omitempty"`
	IndexName          string           `json:"index_name,omitempty"`
	Constraint         *ConstraintState `json:"constraint,omitempty"`
	ConstraintName     string           `json:"constraint_name,omitempty"`
	SQL                string           `json:"sql,omitempty"`
	ReverseSQL         string           `json:"reverse_sql,omitempty"`
	Elidable           bool             `json:"elidable,omitempty"`
	DatabaseOperations []OperationSpec  `json:"database_operations,omitempty"`
	StateOperations    []OperationSpec  `json:"state_operations,omitempty"`
	HasDefault         bool             `json:"has_default,omitempty"`
	UnsafeAcknowledged bool             `json:"unsafe_acknowledged,omitempty"`
}

// OperationSpecProvider is implemented by operations that can serialize their
// full migration content into a stable manifest.
type OperationSpecProvider interface {
	MigrationOperationSpec() OperationSpec
}

// OperationSpecFor returns canonical content for an operation.
func OperationSpecFor(operation Operation) OperationSpec {
	if provider, ok := operation.(OperationSpecProvider); ok {
		spec := provider.MigrationOperationSpec()
		if spec.Type == "" {
			spec.Type = operation.Name()
		}
		return spec
	}
	return OperationSpec{Type: operation.Name()}
}

// CanonicalJSON returns deterministic JSON for hashing and generated files.
func (s OperationSpec) CanonicalJSON() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// OperationSpecFromJSON decodes a canonical operation spec.
func OperationSpecFromJSON(data string) (OperationSpec, error) {
	var spec OperationSpec
	if err := json.Unmarshal([]byte(data), &spec); err != nil {
		return OperationSpec{}, err
	}
	return spec, nil
}
