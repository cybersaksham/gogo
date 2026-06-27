package models

import (
	"fmt"

	modelconstraints "github.com/cybersaksham/gogo/models/constraints"
)

const (
	ConstraintUnique = modelconstraints.TypeUnique
	ConstraintCheck  = modelconstraints.TypeCheck
)

// Index describes model index metadata.
type Index = modelconstraints.Index

// IndexField describes one ordered model index field.
type IndexField = modelconstraints.IndexField

// Constraint describes model constraint metadata.
type Constraint = modelconstraints.Constraint

// Asc creates ascending index field metadata.
func Asc(name string) IndexField {
	return modelconstraints.Asc(name)
}

// Desc creates descending index field metadata.
func Desc(name string) IndexField {
	return modelconstraints.Desc(name)
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
		copied[i] = index.Clone()
	}
	return copied
}

func cloneConstraints(constraints []Constraint) []Constraint {
	copied := make([]Constraint, len(constraints))
	for i, constraint := range constraints {
		copied[i] = constraint.Clone()
	}
	return copied
}
