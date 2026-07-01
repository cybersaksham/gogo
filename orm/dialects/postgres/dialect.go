package postgres

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cybersaksham/gogo/orm/dialects"
)

// Dialect renders PostgreSQL SQL syntax.
type Dialect struct{}

// New returns a PostgreSQL dialect.
func New() Dialect {
	return Dialect{}
}

func (Dialect) Name() string {
	return "postgres"
}

func (Dialect) Placeholder(position int) string {
	return "$" + strconv.Itoa(position)
}

func (Dialect) QuoteIdent(identifier string) string {
	return dialects.QuoteQualifiedIdent(identifier)
}

func (Dialect) ColumnType(kind string) (string, bool) {
	columnTypes := map[string]string{
		"auto":       "bigserial",
		"integer":    "integer",
		"bigint":     "bigint",
		"decimal":    "numeric",
		"float":      "double precision",
		"boolean":    "boolean",
		"char":       "varchar",
		"text":       "text",
		"date":       "date",
		"datetime":   "timestamp with time zone",
		"time":       "time",
		"duration":   "interval",
		"binary":     "bytea",
		"json":       "jsonb",
		"uuid":       "uuid",
		"ip_address": "inet",
	}
	value, ok := columnTypes[kind]
	return value, ok
}

func (Dialect) SupportsReturning() bool {
	return true
}

func (Dialect) SupportsUpsert() bool {
	return true
}

func (Dialect) JSONLookup(column string, path []string) (string, error) {
	if err := dialects.ValidateJSONPath(path); err != nil {
		return "", err
	}
	return column + " #> '{" + strings.Join(path, ",") + "}'", nil
}

func (Dialect) DateExtract(part, expression string) (string, error) {
	normalized, err := dialects.ValidateDatePart(part)
	if err != nil {
		return "", err
	}
	if normalized == "weekday" {
		normalized = "dow"
	}
	return "EXTRACT(" + strings.ToUpper(normalized) + " FROM " + expression + ")", nil
}

func (d Dialect) LockClause(options dialects.LockOptions) (string, error) {
	if !options.ForUpdate {
		return "", nil
	}
	if options.NoWait && options.SkipLocked {
		return "", fmt.Errorf("%w: NOWAIT and SKIP LOCKED cannot be combined", dialects.ErrInvalidInput)
	}
	parts := []string{"FOR UPDATE"}
	if len(options.Of) > 0 {
		quoted := make([]string, len(options.Of))
		for i, identifier := range options.Of {
			quoted[i] = d.QuoteIdent(identifier)
		}
		parts = append(parts, "OF "+strings.Join(quoted, ", "))
	}
	if options.NoWait {
		parts = append(parts, "NOWAIT")
	}
	if options.SkipLocked {
		parts = append(parts, "SKIP LOCKED")
	}
	return strings.Join(parts, " "), nil
}

func (Dialect) LimitOffset(options dialects.LimitOffset) string {
	return dialects.RenderLimitOffset(options)
}

func (d Dialect) SavepointSQL(name string) string {
	return "SAVEPOINT " + d.QuoteIdent(name)
}

func (d Dialect) RollbackToSavepointSQL(name string) string {
	return "ROLLBACK TO SAVEPOINT " + d.QuoteIdent(name)
}

func (d Dialect) ReleaseSavepointSQL(name string) string {
	return "RELEASE SAVEPOINT " + d.QuoteIdent(name)
}

func (Dialect) SchemaIntrospection() dialects.SchemaIntrospection {
	return dialects.SchemaIntrospection{
		TablesSQL:      "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema')",
		ColumnsSQL:     "SELECT c.table_name, c.column_name, c.data_type, c.is_nullable = 'YES' AS nullable, COALESCE(pk.primary_key, false) AS primary_key, c.ordinal_position FROM information_schema.columns c LEFT JOIN (SELECT kcu.table_schema, kcu.table_name, kcu.column_name, true AS primary_key FROM information_schema.table_constraints tc JOIN information_schema.key_column_usage kcu ON tc.constraint_schema = kcu.constraint_schema AND tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema AND tc.table_name = kcu.table_name WHERE tc.constraint_type = 'PRIMARY KEY') pk ON c.table_schema = pk.table_schema AND c.table_name = pk.table_name AND c.column_name = pk.column_name WHERE c.table_schema NOT IN ('pg_catalog', 'information_schema') ORDER BY c.table_name, c.ordinal_position",
		ConstraintsSQL: "SELECT conname FROM pg_constraint",
		IndexesSQL:     "SELECT indexname FROM pg_indexes",
	}
}
