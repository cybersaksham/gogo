package models

import (
	"fmt"
	"strings"

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
	if err := validateFields(meta); err != nil {
		return err
	}
	if err := validateIndexes(meta); err != nil {
		return err
	}
	if err := validateConstraints(meta); err != nil {
		return err
	}
	if err := validatePermissions(meta); err != nil {
		return err
	}
	return nil
}

func validateFields(meta Metadata) error {
	if len(meta.Fields) == 0 {
		return nil
	}
	fieldNames := map[string]struct{}{}
	columns := map[string]struct{}{}
	hasPrimaryKey := false
	for _, field := range meta.Fields {
		if strings.TrimSpace(field.Name) == "" {
			return fmt.Errorf("%w: field name is required", ErrInvalidMetadata)
		}
		if _, exists := fieldNames[field.Name]; exists {
			return fmt.Errorf("%w: duplicate field %s", ErrInvalidMetadata, field.Name)
		}
		fieldNames[field.Name] = struct{}{}
		column := field.Column
		if column == "" {
			column = field.Name
		}
		if strings.TrimSpace(column) == "" {
			return fmt.Errorf("%w: column is required for field %s", ErrInvalidMetadata, field.Name)
		}
		if _, exists := columns[column]; exists {
			return fmt.Errorf("%w: duplicate column %s", ErrInvalidMetadata, column)
		}
		columns[column] = struct{}{}
		if field.PrimaryKey {
			hasPrimaryKey = true
		}
		if _, err := NormalizeDatabaseDefault(field.DBDefault); err != nil {
			return fmt.Errorf("%w: field %s has invalid database default: %v", ErrInvalidMetadata, field.Name, err)
		}
	}
	if meta.CompositePrimaryKey != nil {
		if len(meta.CompositePrimaryKey.Columns) == 0 {
			return fmt.Errorf("%w: composite primary key requires columns", ErrInvalidMetadata)
		}
		for _, column := range meta.CompositePrimaryKey.Columns {
			if _, exists := columns[column]; !exists {
				if _, exists := fieldNames[column]; !exists {
					return fmt.Errorf("%w: composite primary key references unknown column %s", ErrInvalidMetadata, column)
				}
			}
		}
		hasPrimaryKey = true
	}
	if !hasPrimaryKey {
		return fmt.Errorf("%w: model %s requires a primary key", ErrInvalidMetadata, meta.Label())
	}
	return nil
}

func validateIndexes(meta Metadata) error {
	seen := map[string]struct{}{}
	for _, index := range meta.Indexes {
		if err := index.Validate(); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidMetadata, err)
		}
		name := index.NameFor(meta.TableName)
		if _, exists := seen[name]; exists {
			return fmt.Errorf("%w: duplicate index %s", ErrInvalidMetadata, name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func validateConstraints(meta Metadata) error {
	seen := map[string]struct{}{}
	for _, constraint := range meta.Constraints {
		if err := constraint.Validate(); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidMetadata, err)
		}
		name := constraint.NameFor(meta.TableName)
		if _, exists := seen[name]; exists {
			return fmt.Errorf("%w: duplicate constraint %s", ErrInvalidMetadata, name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func validatePermissions(meta Metadata) error {
	seen := map[string]struct{}{}
	for _, permission := range meta.Permissions {
		if strings.TrimSpace(permission.CodeName) == "" {
			return fmt.Errorf("%w: permission codename is required", ErrInvalidMetadata)
		}
		if _, exists := seen[permission.CodeName]; exists {
			return fmt.Errorf("%w: duplicate permission %s", ErrInvalidMetadata, permission.CodeName)
		}
		seen[permission.CodeName] = struct{}{}
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
