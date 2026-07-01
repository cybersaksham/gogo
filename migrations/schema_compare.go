package migrations

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cybersaksham/gogo/models"
)

// SchemaDifference describes one mismatch between expected and actual table shape.
type SchemaDifference struct {
	Kind     string
	Table    string
	Column   string
	Expected string
	Actual   string
}

func (d SchemaDifference) String() string {
	target := d.Table
	if d.Column != "" {
		target += "." + d.Column
	}
	if d.Expected != "" || d.Actual != "" {
		return fmt.Sprintf("%s mismatch %s expected %s actual %s", d.Kind, target, d.Expected, d.Actual)
	}
	return fmt.Sprintf("%s %s", d.Kind, target)
}

// CompareTableSchema compares an expected initial table shape with inspected columns.
func CompareTableSchema(expected TableSchema, actualColumns []ColumnSchema) []SchemaDifference {
	actualByName := make(map[string]ColumnSchema, len(actualColumns))
	for _, column := range actualColumns {
		actualByName[strings.ToLower(column.Name)] = column
	}
	var diffs []SchemaDifference
	for _, column := range expected.Columns {
		actual, ok := actualByName[strings.ToLower(column.Name)]
		if !ok {
			diffs = append(diffs, SchemaDifference{Kind: "MISSING column", Table: expected.Name, Column: column.Name})
			continue
		}
		if column.PrimaryKey && !actual.PrimaryKey {
			diffs = append(diffs, SchemaDifference{Kind: "PRIMARY KEY", Table: expected.Name, Column: column.Name, Expected: "primary key", Actual: "not primary key"})
		}
		if !column.Nullable && actual.Nullable && !column.PrimaryKey {
			diffs = append(diffs, SchemaDifference{Kind: "NULL", Table: expected.Name, Column: column.Name, Expected: "not null", Actual: "nullable"})
		}
		expectedKind := comparableColumnKind(column)
		actualKind := comparableColumnKind(actual)
		if expectedKind != "" && actualKind != "" && expectedKind != actualKind {
			diffs = append(diffs, SchemaDifference{Kind: "TYPE", Table: expected.Name, Column: column.Name, Expected: expectedKind, Actual: actualKind})
		}
		if (column.Default != nil || actual.Default != nil || strings.TrimSpace(actual.DefaultSQL) != "") && !databaseDefaultMatches(column.Default, actual.Default, actual.DefaultSQL) {
			diffs = append(diffs, SchemaDifference{Kind: "DEFAULT", Table: expected.Name, Column: column.Name, Expected: normalizeDefaultSQL(renderDatabaseDefault(column.Default)), Actual: normalizeDefaultSQL(actual.DefaultSQL)})
		}
		if column.Collation != "" && column.Collation != actual.Collation {
			diffs = append(diffs, SchemaDifference{Kind: "COLLATION", Table: expected.Name, Column: column.Name, Expected: column.Collation, Actual: actual.Collation})
		}
	}
	return diffs
}

func comparableColumnKind(column ColumnSchema) string {
	if column.NormalizedKind != "" {
		return strings.ToLower(strings.TrimSpace(column.NormalizedKind))
	}
	return NormalizeColumnKind(column.Kind)
}

// NormalizeColumnKind lowercases and trims a database type for shape comparison.
func NormalizeColumnKind(kind string) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	kind = strings.Join(strings.Fields(kind), " ")
	return kind
}

func databaseDefaultMatches(expected, actual *models.DatabaseDefault, actualSQL string) bool {
	expectedSQL := normalizeDefaultSQL(renderDatabaseDefault(expected))
	if actual != nil {
		return expectedSQL == normalizeDefaultSQL(renderDatabaseDefault(actual))
	}
	return expectedSQL == normalizeDefaultSQL(actualSQL)
}

func renderDatabaseDefault(defaultValue *models.DatabaseDefault) string {
	if defaultValue == nil {
		return ""
	}
	normalized, err := models.NormalizeDatabaseDefault(defaultValue)
	if err != nil {
		return ""
	}
	switch normalized.Kind {
	case models.DefaultExpression:
		return normalized.SQL
	case models.DefaultLiteral:
		return renderLiteralDefault(normalized.Value)
	default:
		return ""
	}
}

func renderLiteralDefault(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case string:
		return "'" + strings.ReplaceAll(typed, "'", "''") + "'"
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case int:
		return strconv.FormatInt(int64(typed), 10)
	case int8:
		return strconv.FormatInt(int64(typed), 10)
	case int16:
		return strconv.FormatInt(int64(typed), 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint8:
		return strconv.FormatUint(uint64(typed), 10)
	case uint16:
		return strconv.FormatUint(uint64(typed), 10)
	case uint32:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func normalizeDefaultSQL(sql string) string {
	sql = strings.TrimSpace(sql)
	for strings.HasPrefix(sql, "(") && strings.HasSuffix(sql, ")") {
		sql = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(sql, "("), ")"))
	}
	if index := strings.Index(sql, "::"); index > 0 {
		sql = sql[:index]
	}
	return strings.ToLower(strings.TrimSpace(sql))
}
