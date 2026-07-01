package schema

import (
	"reflect"
	"testing"

	"github.com/cybersaksham/gogo/migrations"
	"github.com/cybersaksham/gogo/models"
	"github.com/cybersaksham/gogo/orm/dialects/postgres"
	"github.com/cybersaksham/gogo/orm/dialects/sqlite"
)

func TestSchemaEditorPostgresGoldenSQL(t *testing.T) {
	editor := NewEditor(postgres.New())
	sql := []string{
		editor.CreateTable("blog_post", []migrations.FieldState{{Name: "id", Kind: "bigserial", PrimaryKey: true}, {Name: "title", Kind: "text", Null: false}}),
		editor.RenameTable("blog_post", "blog_article"),
		editor.AddColumn("blog_article", migrations.FieldState{Name: "slug", Kind: "text", Null: true}),
		editor.AlterColumnType("blog_article", "slug", "varchar(255)"),
		editor.AlterNull("blog_article", "slug", false),
		editor.AlterDefault("blog_article", "slug", models.DefaultValue("draft")),
		editor.RenameColumn("blog_article", "slug", "url_slug"),
		editor.AddIndex("blog_article", migrations.IndexState{Name: "idx_slug", Fields: []string{"url_slug"}}),
		editor.RenameIndex("idx_slug", "idx_url_slug"),
		editor.AddConstraint("blog_article", migrations.ConstraintState{Name: "uniq_slug", Type: "unique", Fields: []string{"url_slug"}}),
		editor.CreateManyToManyTable("blog_post_tags", "post_id", "tag_id"),
	}
	want := []string{
		`CREATE TABLE "blog_post" ("id" bigserial PRIMARY KEY NOT NULL, "title" text NOT NULL)`,
		`ALTER TABLE "blog_post" RENAME TO "blog_article"`,
		`ALTER TABLE "blog_article" ADD COLUMN "slug" text`,
		`ALTER TABLE "blog_article" ALTER COLUMN "slug" TYPE varchar(255)`,
		`ALTER TABLE "blog_article" ALTER COLUMN "slug" SET NOT NULL`,
		`ALTER TABLE "blog_article" ALTER COLUMN "slug" SET DEFAULT 'draft'`,
		`ALTER TABLE "blog_article" RENAME COLUMN "slug" TO "url_slug"`,
		`CREATE INDEX "idx_slug" ON "blog_article" ("url_slug")`,
		`ALTER INDEX "idx_slug" RENAME TO "idx_url_slug"`,
		`ALTER TABLE "blog_article" ADD CONSTRAINT "uniq_slug" UNIQUE ("url_slug")`,
		`CREATE TABLE "blog_post_tags" ("post_id" bigint NOT NULL, "tag_id" bigint NOT NULL)`,
	}
	if !reflect.DeepEqual(sql, want) {
		t.Fatalf("postgres sql = %#v", sql)
	}
}

func TestSchemaEditorRendersDatabaseDefaults(t *testing.T) {
	literal := models.DefaultValue("draft's")
	expression := models.DefaultSQL("gen_random_uuid()")
	editor := NewEditor(postgres.New())

	got := []string{
		editor.CreateTable("blog_post", []migrations.FieldState{
			{Name: "id", Column: "id", Kind: "uuid", PrimaryKey: true, DBDefault: &expression},
			{Name: "status", Column: "status", Kind: "text", DBDefault: &literal},
		}),
		editor.AddColumn("blog_post", migrations.FieldState{Name: "visible", Kind: "boolean", DBDefault: databaseDefaultPtr(models.DefaultValue(true))}),
		editor.AlterDefault("blog_post", "status", literal),
		editor.AlterDefault("blog_post", "removed", nil),
	}
	want := []string{
		`CREATE TABLE "blog_post" ("id" uuid DEFAULT gen_random_uuid() PRIMARY KEY NOT NULL, "status" text DEFAULT 'draft''s' NOT NULL)`,
		`ALTER TABLE "blog_post" ADD COLUMN "visible" boolean DEFAULT true NOT NULL`,
		`ALTER TABLE "blog_post" ALTER COLUMN "status" SET DEFAULT 'draft''s'`,
		`ALTER TABLE "blog_post" ALTER COLUMN "removed" DROP DEFAULT`,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("default SQL = %#v", got)
	}
}

func TestSchemaEditorSQLiteGoldenSQL(t *testing.T) {
	editor := NewEditor(sqlite.New())
	if got := editor.DropColumn("blog_post", "slug"); got != `ALTER TABLE "blog_post" DROP COLUMN "slug"` {
		t.Fatalf("DropColumn() = %q", got)
	}
	if got := editor.DropIndex("idx_slug"); got != `DROP INDEX "idx_slug"` {
		t.Fatalf("DropIndex() = %q", got)
	}
	if got := editor.AddConstraint("blog_post", migrations.ConstraintState{Name: "uniq_slug", Type: "unique", Fields: []string{"slug"}}); got != `-- SQLite rebuild required to add constraint "uniq_slug" on "blog_post"` {
		t.Fatalf("AddConstraint() = %q", got)
	}
	if got := editor.DropConstraint("blog_post", "uniq_slug"); got != `-- SQLite rebuild required to drop constraint "uniq_slug" on "blog_post"` {
		t.Fatalf("DropConstraint() = %q", got)
	}
	if got := editor.DropManyToManyTable("blog_post_tags"); got != `DROP TABLE "blog_post_tags"` {
		t.Fatalf("DropManyToManyTable() = %q", got)
	}
}

func databaseDefaultPtr(defaultValue models.DatabaseDefault) *models.DatabaseDefault {
	return &defaultValue
}
