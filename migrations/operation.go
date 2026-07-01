package migrations

import (
	"context"
	"strings"
)

// SchemaEditor is the database-facing contract used by operations.
type SchemaEditor interface {
	Execute(context.Context, string, ...any) error
}

// TableExistenceChecker is implemented by schema editors that can inspect existing tables.
type TableExistenceChecker interface {
	TableExists(context.Context, string) (bool, error)
}

// InitialTableProvider lets fake-initial compare an initial migration with existing schema.
type InitialTableProvider interface {
	InitialTables() []string
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

// InitialTableNameFromSQL returns the table created by a simple CREATE TABLE statement.
func InitialTableNameFromSQL(statement string) (string, bool) {
	fields := strings.Fields(strings.TrimSpace(statement))
	if len(fields) < 3 || !strings.EqualFold(fields[0], "CREATE") || !strings.EqualFold(fields[1], "TABLE") {
		return "", false
	}
	index := 2
	if len(fields) > 5 && strings.EqualFold(fields[index], "IF") && strings.EqualFold(fields[index+1], "NOT") && strings.EqualFold(fields[index+2], "EXISTS") {
		index += 3
	}
	if index >= len(fields) {
		return "", false
	}
	name := fields[index]
	if paren := strings.Index(name, "("); paren >= 0 {
		name = name[:paren]
	}
	name = strings.Trim(name, "`\"[]")
	if name == "" {
		return "", false
	}
	return name, true
}
