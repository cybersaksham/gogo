package postgres

import (
	"strings"
	"testing"

	"github.com/cybersaksham/gogo/orm/dialects"
)

func TestPostgresDialectRendering(t *testing.T) {
	dialect := New()
	if dialect.Name() != "postgres" {
		t.Fatalf("Name() = %q", dialect.Name())
	}
	if got := dialect.Placeholder(3); got != "$3" {
		t.Fatalf("Placeholder(3) = %q", got)
	}
	if got := dialect.QuoteIdent(`user"name`); got != `"user""name"` {
		t.Fatalf("QuoteIdent() = %q", got)
	}
	if !dialect.SupportsReturning() || !dialect.SupportsUpsert() {
		t.Fatalf("returning/upsert support not enabled")
	}
	if got := dialect.LimitOffset(dialects.LimitOffset{Limit: dialects.Int(10), Offset: dialects.Int(5)}); got != "LIMIT 10 OFFSET 5" {
		t.Fatalf("LimitOffset() = %q", got)
	}
	lock, err := dialect.LockClause(dialects.LockOptions{ForUpdate: true, Of: []string{"blog_post"}, SkipLocked: true})
	if err != nil {
		t.Fatalf("LockClause() error = %v", err)
	}
	if lock != `FOR UPDATE OF "blog_post" SKIP LOCKED` {
		t.Fatalf("LockClause() = %q", lock)
	}
	if got, ok := dialect.ColumnType("json"); !ok || got != "jsonb" {
		t.Fatalf("ColumnType(json) = (%q, %v)", got, ok)
	}
}

func TestPostgresDateJSONAndSavepoints(t *testing.T) {
	dialect := New()
	dateSQL, err := dialect.DateExtract("year", `"created_at"`)
	if err != nil {
		t.Fatalf("DateExtract() error = %v", err)
	}
	if dateSQL != `EXTRACT(YEAR FROM "created_at")` {
		t.Fatalf("DateExtract() = %q", dateSQL)
	}
	jsonSQL, err := dialect.JSONLookup(`"data"`, []string{"owner", "email"})
	if err != nil {
		t.Fatalf("JSONLookup() error = %v", err)
	}
	if jsonSQL != `"data" #> '{owner,email}'` {
		t.Fatalf("JSONLookup() = %q", jsonSQL)
	}
	if got := dialect.SavepointSQL("sp1"); got != `SAVEPOINT "sp1"` {
		t.Fatalf("SavepointSQL() = %q", got)
	}
	if got := dialect.ReleaseSavepointSQL("sp1"); got != `RELEASE SAVEPOINT "sp1"` {
		t.Fatalf("ReleaseSavepointSQL() = %q", got)
	}
}

func TestPostgresDialectColumnShapeIntrospectionSQL(t *testing.T) {
	sql := New().SchemaIntrospection().ColumnsSQL
	if sql == "" {
		t.Fatal("ColumnsSQL is empty")
	}
	for _, want := range []string{"table_schema", "table_name", "column_name", "data_type", "udt_name", "character_maximum_length", "numeric_precision", "numeric_scale", "column_default", "collation_name", "is_identity", "primary_key", "ordinal_position"} {
		if !strings.Contains(sql, want) {
			t.Fatalf("ColumnsSQL missing %q: %s", want, sql)
		}
	}
}
