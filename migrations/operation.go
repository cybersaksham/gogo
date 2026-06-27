package migrations

import "context"

// SchemaEditor is the database-facing contract used by operations.
type SchemaEditor interface {
	Execute(context.Context, string, ...any) error
}

// Operation is the complete migration operation contract.
type Operation interface {
	Name() string
	StateForwards(*ProjectState) error
	DatabaseForwards(context.Context, SchemaEditor) error
	DatabaseBackwards(context.Context, SchemaEditor) error
	Describe() string
	Reversible() bool
	ReferencesModel(appLabel, modelName string) bool
	ReferencesField(appLabel, modelName, fieldName string) bool
}
