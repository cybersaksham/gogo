package migrations

import (
	"context"
	"strings"
)

// SchemaEditor is the database-facing contract used by operations.
type SchemaEditor interface {
	Execute(context.Context, string, ...any) error
}

// SchemaRenderer optionally centralizes dialect-specific schema SQL rendering.
type SchemaRenderer interface {
	CreateTable(table string, fields []FieldState) string
	DropTable(table string) string
	RenameTable(oldName, newName string) string
	AddColumn(table string, field FieldState) string
	DropColumn(table, column string) string
	AlterColumnType(table, column, kind string) string
	RenameColumn(table, oldName, newName string) string
	AddIndex(table string, index IndexState) string
	DropIndex(name string) string
	RenameIndex(oldName, newName string) string
	AddConstraint(table string, constraint ConstraintState) string
	DropConstraint(table, name string) string
}

// TableExistenceChecker is implemented by schema editors that can inspect existing tables.
type TableExistenceChecker interface {
	TableExists(context.Context, string) (bool, error)
}

// TableShapeChecker is implemented by schema editors that can inspect columns.
type TableShapeChecker interface {
	TableColumns(context.Context, string) ([]ColumnSchema, error)
}

// InitialTableProvider lets fake-initial compare an initial migration with existing schema.
type InitialTableProvider interface {
	InitialTables() []string
}

// InitialSchemaProvider lets fake-initial validate existing table shape.
type InitialSchemaProvider interface {
	InitialSchema() []TableSchema
}

// TableSchema describes the minimum table shape required by an initial migration.
type TableSchema struct {
	Name    string
	Columns []ColumnSchema
}

// ColumnSchema describes one column required by an initial migration.
type ColumnSchema struct {
	Name            string
	Kind            string
	PrimaryKey      bool
	Nullable        bool
	OrdinalPosition int
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
