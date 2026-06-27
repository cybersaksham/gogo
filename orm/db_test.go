package orm

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"
)

type testDialect struct{}

func (testDialect) Name() string { return "sqlite" }

func TestOpenDatabasePingStatsAndClose(t *testing.T) {
	ctx := context.Background()
	database, err := OpenDatabase(ctx, DatabaseConfig{
		Name:    DefaultDatabase,
		Driver:  "sqlite",
		DSN:     ":memory:",
		Dialect: testDialect{},
	})
	if err != nil {
		t.Fatalf("OpenDatabase() error = %v", err)
	}

	if database.Name != DefaultDatabase || database.Driver != "sqlite" || database.Dialect.Name() != "sqlite" {
		t.Fatalf("database metadata = %#v", database)
	}
	if database.SQLDB() == nil {
		t.Fatalf("SQLDB() is nil")
	}
	if err := database.Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if stats := database.Stats(); stats.MaxOpenConnections < 0 {
		t.Fatalf("Stats() = %#v", stats)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestOpenDatabaseUsesHealthCheck(t *testing.T) {
	called := false
	_, err := OpenDatabase(context.Background(), DatabaseConfig{
		Name:    "health",
		Driver:  "sqlite",
		DSN:     ":memory:",
		Dialect: testDialect{},
		HealthCheck: func(ctx context.Context, db *sql.DB) error {
			called = true
			return db.PingContext(ctx)
		},
	})
	if err != nil {
		t.Fatalf("OpenDatabase() error = %v", err)
	}
	if !called {
		t.Fatalf("health check was not called")
	}
}

func TestOpenDatabasePropagatesHealthCheckFailure(t *testing.T) {
	failure := errors.New("database unavailable")
	_, err := OpenDatabase(context.Background(), DatabaseConfig{
		Name:    "bad",
		Driver:  "sqlite",
		DSN:     ":memory:",
		Dialect: testDialect{},
		HealthCheck: func(context.Context, *sql.DB) error {
			return failure
		},
	})
	if !errors.Is(err, failure) {
		t.Fatalf("OpenDatabase() error = %v, want health failure", err)
	}
}

func TestConnectionsSupportDefaultAndNamedDatabases(t *testing.T) {
	ctx := context.Background()
	defaultDB, err := OpenDatabase(ctx, DatabaseConfig{Name: DefaultDatabase, Driver: "sqlite", DSN: ":memory:", Dialect: testDialect{}})
	if err != nil {
		t.Fatalf("Open default database: %v", err)
	}
	replicaDB, err := OpenDatabase(ctx, DatabaseConfig{Name: "replica", Driver: "sqlite", DSN: ":memory:", Dialect: testDialect{}})
	if err != nil {
		t.Fatalf("Open replica database: %v", err)
	}
	defer defaultDB.Close()
	defer replicaDB.Close()

	connections := NewConnections()
	if err := connections.Add(defaultDB); err != nil {
		t.Fatalf("Add(default) error = %v", err)
	}
	if err := connections.Add(replicaDB); err != nil {
		t.Fatalf("Add(replica) error = %v", err)
	}

	gotDefault, ok := connections.Default()
	if !ok || gotDefault.Name != DefaultDatabase {
		t.Fatalf("Default() = (%#v, %v)", gotDefault, ok)
	}
	gotReplica, ok := connections.Get("replica")
	if !ok || gotReplica.Name != "replica" {
		t.Fatalf("Get(replica) = (%#v, %v)", gotReplica, ok)
	}
	if err := connections.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestConnectionsRejectDuplicatesAndMissingDefaults(t *testing.T) {
	database := &Database{Name: "replica"}
	connections := NewConnections()
	if err := connections.Add(database); err != nil {
		t.Fatalf("Add(replica) error = %v", err)
	}
	if err := connections.Add(database); !errors.Is(err, ErrDatabaseExists) {
		t.Fatalf("Add(duplicate) error = %v, want ErrDatabaseExists", err)
	}
	if _, ok := connections.Default(); ok {
		t.Fatalf("Default() ok = true without default database")
	}
}
