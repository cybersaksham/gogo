package migrations

import (
	"context"
	"testing"

	"github.com/cybersaksham/gogo/orm"
	sqlitedialect "github.com/cybersaksham/gogo/orm/dialects/sqlite"
	_ "modernc.org/sqlite"
)

func TestRecorderAppliedHistory(t *testing.T) {
	ctx := context.Background()
	database := openRecorderDB(t)
	defer database.Close()
	recorder := NewRecorder(database, "test-executor")
	if err := recorder.EnsureSchema(ctx); err != nil {
		t.Fatalf("EnsureSchema() error = %v", err)
	}
	migration := testMigration("blog", "0001_initial")
	if ok, err := recorder.IsApplied(ctx, migration.Dependency()); err != nil || ok {
		t.Fatalf("IsApplied(before) = (%v, %v)", ok, err)
	}
	if err := recorder.RecordApplied(ctx, migration, "abc123"); err != nil {
		t.Fatalf("RecordApplied() error = %v", err)
	}
	if ok, err := recorder.IsApplied(ctx, migration.Dependency()); err != nil || !ok {
		t.Fatalf("IsApplied(after) = (%v, %v)", ok, err)
	}
	applied, err := recorder.Applied(ctx)
	if err != nil {
		t.Fatalf("Applied() error = %v", err)
	}
	if len(applied) != 1 || applied[0].Checksum != "abc123" || applied[0].ExecutorVersion != "test-executor" {
		t.Fatalf("applied = %#v", applied)
	}
	if err := recorder.CheckConsistency(ctx, []Migration{migration}); err != nil {
		t.Fatalf("CheckConsistency() error = %v", err)
	}
	if err := recorder.RecordUnapplied(ctx, migration.Dependency()); err != nil {
		t.Fatalf("RecordUnapplied() error = %v", err)
	}
	if ok, _ := recorder.IsApplied(ctx, migration.Dependency()); ok {
		t.Fatalf("migration still applied")
	}
}

func openRecorderDB(t *testing.T) *orm.Database {
	t.Helper()
	database, err := orm.OpenDatabase(context.Background(), orm.DatabaseConfig{
		Name:    orm.DefaultDatabase,
		Driver:  "sqlite",
		DSN:     ":memory:",
		Dialect: sqlitedialect.New(),
	})
	if err != nil {
		t.Fatalf("OpenDatabase() error = %v", err)
	}
	return database
}
