package testing

import (
	"context"
	"fmt"
	"reflect"
	stdtesting "testing"
)

func TestFixtureLoadingDumpingFactoriesAndNaturalKeys(t *stdtesting.T) {
	ctx := context.Background()
	database, err := NewSQLiteTestDatabase(ctx)
	if err != nil {
		t.Fatalf("NewSQLiteTestDatabase() error = %v", err)
	}
	defer database.Close()
	if _, err := database.SQLDB().ExecContext(ctx, `CREATE TABLE blog_post (id INTEGER PRIMARY KEY, title TEXT NOT NULL, slug TEXT NOT NULL)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	factory := NewFactory("blog.Post", map[string]any{
		"id":    Seq(func(n int) any { return n }),
		"title": Seq(func(n int) any { return fmt.Sprintf("Post %d", n) }),
		"slug":  Seq(func(n int) any { return fmt.Sprintf("post-%d", n) }),
	}).WithNaturalKey("slug")

	fixtures := factory.BuildMany(2)
	if !reflect.DeepEqual(fixtures[0].NaturalKey, []any{"post-1"}) || fixtures[1].PK != 2 {
		t.Fatalf("factory fixtures = %#v", fixtures)
	}
	if err := LoadFixtures(ctx, database, fixtures); err != nil {
		t.Fatalf("LoadFixtures() error = %v", err)
	}

	dumped, err := DumpFixtures(ctx, database, "blog.Post")
	if err != nil {
		t.Fatalf("DumpFixtures() error = %v", err)
	}
	if len(dumped) != 2 || dumped[0].Model != "blog.Post" || dumped[0].PK != int64(1) || dumped[0].Fields["title"] != "Post 1" {
		t.Fatalf("DumpFixtures() = %#v", dumped)
	}
}

func TestTransactionalFixturesRollbackAfterCallback(t *stdtesting.T) {
	ctx := context.Background()
	database, err := NewSQLiteTestDatabase(ctx)
	if err != nil {
		t.Fatalf("NewSQLiteTestDatabase() error = %v", err)
	}
	defer database.Close()
	if _, err := database.SQLDB().ExecContext(ctx, `CREATE TABLE blog_post (id INTEGER PRIMARY KEY, title TEXT NOT NULL)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	fixtures := []Fixture{{Model: "blog.Post", PK: 1, Fields: map[string]any{"title": "Temporary"}}}
	if err := TransactionalFixtures(ctx, database, fixtures, func(ctx context.Context, tx Transaction) error {
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM blog_post`).Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			t.Fatalf("fixture count inside transaction = %d, want 1", count)
		}
		return nil
	}); err != nil {
		t.Fatalf("TransactionalFixtures() error = %v", err)
	}

	if count := countRows(t, database, "blog_post"); count != 0 {
		t.Fatalf("fixture count after rollback = %d, want 0", count)
	}
}
