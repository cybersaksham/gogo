package cli

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cybersaksham/gogo/models"

	_ "modernc.org/sqlite"
)

func TestInspectDBPrintsExistingSchemaMetadata(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "db.sqlite3")
	writeSchemaTestEnv(t, dir, dbPath)
	db := openSchemaTestDB(t, dbPath)
	if _, err := db.Exec(`CREATE TABLE legacy_item (id integer PRIMARY KEY, name text NOT NULL)`); err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	db.Close()
	t.Chdir(dir)

	var stdout bytes.Buffer
	if err := NewRoot().Execute(context.Background(), []string{"inspectdb", "--table", "legacy_item"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("inspectdb error = %v", err)
	}
	output := stdout.String()
	for _, want := range []string{`ModelName: "LegacyItem"`, `DBTable: "legacy_item"`, `Managed: &legacyItemManaged`, `field id INTEGER primary_key`, `field name TEXT`} {
		if !strings.Contains(output, want) {
			t.Fatalf("inspectdb output missing %q:\n%s", want, output)
		}
	}
}

func TestDiffSchemaReportsMissingColumnsAndPassesMatch(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "db.sqlite3")
	writeSchemaTestEnv(t, dir, dbPath)
	db := openSchemaTestDB(t, dbPath)
	if _, err := db.Exec(`CREATE TABLE blog_item (id integer PRIMARY KEY, name text NOT NULL)`); err != nil {
		t.Fatalf("create blog table: %v", err)
	}
	db.Close()
	t.Chdir(dir)

	meta := models.Metadata{
		AppLabel:  "blog",
		ModelName: "Item",
		TableName: "blog_item",
		Fields: []models.FieldMeta{
			{Name: "id", Column: "id", PrimaryKey: true},
			{Name: "name", Column: "name"},
			{Name: "slug", Column: "slug"},
		},
	}
	root := NewRootWithOptions(RootOptions{ProjectModels: []models.Metadata{meta}})
	var stdout bytes.Buffer
	err := root.Execute(context.Background(), []string{"diffschema", "--app", "blog"}, &stdout, &bytes.Buffer{})
	if !errors.Is(err, ErrCommandFailed) {
		t.Fatalf("diffschema error = %v, want ErrCommandFailed", err)
	}
	if !strings.Contains(stdout.String(), "MISSING column blog_item.slug") {
		t.Fatalf("diffschema output = %q", stdout.String())
	}

	db = openSchemaTestDB(t, dbPath)
	if _, err := db.Exec(`ALTER TABLE blog_item ADD COLUMN slug text`); err != nil {
		t.Fatalf("add slug: %v", err)
	}
	db.Close()
	stdout.Reset()
	if err := root.Execute(context.Background(), []string{"diffschema", "--app", "blog"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("diffschema match error = %v\n%s", err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "schema matches model metadata") {
		t.Fatalf("diffschema match output = %q", stdout.String())
	}
}

func writeSchemaTestEnv(t *testing.T, dir, dbPath string) {
	t.Helper()
	writeTextFile(t, filepath.Join(dir, ".env"), "GOGO_SECRET_KEY=schema-secret\nDATABASE_URL=sqlite://"+filepath.ToSlash(dbPath)+"\n")
}

func openSchemaTestDB(t *testing.T, dbPath string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}
