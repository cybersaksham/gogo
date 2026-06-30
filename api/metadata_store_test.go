package api

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/cybersaksham/gogo/models"
	"github.com/cybersaksham/gogo/orm"
	sqlitedialect "github.com/cybersaksham/gogo/orm/dialects/sqlite"

	_ "modernc.org/sqlite"
)

func TestMetadataViewSetStorePersistsCRUD(t *testing.T) {
	ctx := context.Background()
	database, err := orm.OpenDatabase(ctx, orm.DatabaseConfig{
		Name:    orm.DefaultDatabase,
		Driver:  "sqlite",
		DSN:     filepath.Join(t.TempDir(), "api.sqlite3"),
		Dialect: sqlitedialect.New(),
	})
	if err != nil {
		t.Fatalf("OpenDatabase() error = %v", err)
	}
	defer database.Close()
	if _, err := database.SQLDB().ExecContext(ctx, `CREATE TABLE notes_item (id bigint PRIMARY KEY, name text NOT NULL, slug text NOT NULL, created_at timestamp, updated_at timestamp)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	store := NewMetadataViewSetStore(orm.NewMetadataStore(database, apiNotesItemMeta()), apiNotesItemMeta())
	created, err := store.Create(ctx, nil, map[string]any{"name": "API", "slug": "api"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created["id"] != int64(1) || created["name"] != "API" {
		t.Fatalf("created = %#v", created)
	}

	list, err := store.List(ctx, nil)
	if err != nil || len(list) != 1 {
		t.Fatalf("List() = %#v, %v", list, err)
	}

	updated, err := store.Update(ctx, nil, "1", map[string]any{"slug": "api-updated"}, true)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated["name"] != "API" || updated["slug"] != "api-updated" {
		t.Fatalf("updated = %#v", updated)
	}

	if err := store.Destroy(ctx, nil, "1"); err != nil {
		t.Fatalf("Destroy() error = %v", err)
	}
	if _, err := store.Retrieve(ctx, nil, "1"); err != ErrNotFound {
		t.Fatalf("Retrieve(deleted) error = %v, want ErrNotFound", err)
	}
}

func apiNotesItemMeta() models.Metadata {
	return models.Metadata{
		AppLabel:  "notes",
		ModelName: "Item",
		TableName: "notes_item",
		Fields: []models.FieldMeta{
			{Name: "id", Column: "id", PrimaryKey: true},
			{Name: "name", Column: "name"},
			{Name: "slug", Column: "slug"},
			{Name: "created_at", Column: "created_at"},
			{Name: "updated_at", Column: "updated_at"},
		},
	}
}
