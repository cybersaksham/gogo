package dialects

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrUnsupportedFeature = errors.New("unsupported dialect feature")
	ErrInvalidInput       = errors.New("invalid dialect input")
)

// Dialect renders SQL features that differ between database backends.
type Dialect interface {
	Name() string
	Placeholder(position int) string
	QuoteIdent(identifier string) string
	ColumnType(kind string) (string, bool)
	SupportsReturning() bool
	SupportsUpsert() bool
	JSONLookup(column string, path []string) (string, error)
	DateExtract(part, expression string) (string, error)
	LockClause(options LockOptions) (string, error)
	LimitOffset(options LimitOffset) string
	SavepointSQL(name string) string
	RollbackToSavepointSQL(name string) string
	ReleaseSavepointSQL(name string) string
	SchemaIntrospection() SchemaIntrospection
}

// LockOptions describes SELECT ... FOR UPDATE variants.
type LockOptions struct {
	ForUpdate  bool
	NoWait     bool
	SkipLocked bool
	Of         []string
}

// LimitOffset stores optional limit and offset values.
type LimitOffset struct {
	Limit  *int
	Offset *int
}

// SchemaIntrospection stores backend-specific schema inspection SQL hooks.
type SchemaIntrospection struct {
	TablesSQL      string
	ColumnsSQL     string
	ConstraintsSQL string
	IndexesSQL     string
}

// Int returns a pointer for optional integer options.
func Int(value int) *int {
	return &value
}

// QuoteIdent quotes a SQL identifier using ANSI double-quote escaping.
func QuoteIdent(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

// QuoteQualifiedIdent quotes dot-separated identifier parts.
func QuoteQualifiedIdent(identifier string) string {
	parts := strings.Split(identifier, ".")
	for i, part := range parts {
		parts[i] = QuoteIdent(part)
	}
	return strings.Join(parts, ".")
}

// RenderLimitOffset renders portable LIMIT/OFFSET syntax.
func RenderLimitOffset(options LimitOffset) string {
	parts := make([]string, 0, 2)
	if options.Limit != nil {
		parts = append(parts, "LIMIT "+strconv.Itoa(*options.Limit))
	}
	if options.Offset != nil {
		parts = append(parts, "OFFSET "+strconv.Itoa(*options.Offset))
	}
	return strings.Join(parts, " ")
}

// ValidateJSONPath rejects path fragments that cannot be safely rendered as a literal path.
func ValidateJSONPath(path []string) error {
	if len(path) == 0 {
		return fmt.Errorf("%w: JSON path cannot be empty", ErrInvalidInput)
	}
	for _, part := range path {
		if strings.TrimSpace(part) == "" || strings.ContainsAny(part, "'{},") {
			return fmt.Errorf("%w: invalid JSON path part %q", ErrInvalidInput, part)
		}
	}
	return nil
}

// ValidateDatePart normalizes supported date extraction parts.
func ValidateDatePart(part string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(part))
	switch normalized {
	case "year", "quarter", "month", "week", "day", "weekday", "hour", "minute", "second":
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: unsupported date part %q", ErrUnsupportedFeature, part)
	}
}
