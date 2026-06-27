package models

import "fmt"

const (
	ConstraintUnique = "unique"
	ConstraintCheck  = "check"
)

// Index describes model index metadata.
type Index struct {
	Name   string
	Fields []string
}

// Constraint describes model constraint metadata.
type Constraint struct {
	Name   string
	Type   string
	Fields []string
	Check  string
}

// Permission describes one model-level permission.
type Permission struct {
	CodeName string
	Name     string
}

// IsManaged returns true unless Managed was explicitly set to false.
func (m Metadata) IsManaged() bool {
	if m.Managed == nil {
		return true
	}
	return *m.Managed
}

// ValidateMetadata validates model metadata combinations.
func ValidateMetadata(meta Metadata) error {
	if !meta.IsManaged() && meta.GenerateMigrations {
		return fmt.Errorf("%w: unmanaged models cannot generate migrations", ErrInvalidMetadata)
	}
	return nil
}

func cloneIndexes(indexes []Index) []Index {
	copied := make([]Index, len(indexes))
	for i, index := range indexes {
		copied[i] = Index{
			Name:   index.Name,
			Fields: append([]string(nil), index.Fields...),
		}
	}
	return copied
}

func cloneConstraints(constraints []Constraint) []Constraint {
	copied := make([]Constraint, len(constraints))
	for i, constraint := range constraints {
		copied[i] = Constraint{
			Name:   constraint.Name,
			Type:   constraint.Type,
			Fields: append([]string(nil), constraint.Fields...),
			Check:  constraint.Check,
		}
	}
	return copied
}
