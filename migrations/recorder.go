package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cybersaksham/gogo/orm"
)

// AppliedMigration stores one migration history row.
type AppliedMigration struct {
	AppLabel        string
	Name            string
	Applied         time.Time
	Checksum        string
	ExecutorVersion string
}

// Recorder stores migration history in the database.
type Recorder struct {
	Database        *orm.Database
	ExecutorVersion string
}

// NewRecorder creates a migration recorder.
func NewRecorder(database *orm.Database, executorVersion string) Recorder {
	return Recorder{Database: database, ExecutorVersion: executorVersion}
}

// EnsureSchema creates the migration history table.
func (r Recorder) EnsureSchema(ctx context.Context) error {
	_, err := r.db().ExecContext(ctx, `CREATE TABLE IF NOT EXISTS gogo_migrations (app TEXT NOT NULL, name TEXT NOT NULL, applied TIMESTAMP NOT NULL, checksum TEXT NOT NULL, executor_version TEXT NOT NULL, PRIMARY KEY(app, name))`)
	return err
}

// IsApplied reports whether a migration is applied.
func (r Recorder) IsApplied(ctx context.Context, dependency Dependency) (bool, error) {
	var exists int
	err := r.db().QueryRowContext(ctx, `SELECT 1 FROM gogo_migrations WHERE app = ? AND name = ?`, dependency.AppLabel, dependency.Name).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// RecordApplied inserts or replaces an applied migration row.
func (r Recorder) RecordApplied(ctx context.Context, migration Migration, checksum string) error {
	if err := r.EnsureSchema(ctx); err != nil {
		return err
	}
	_, err := r.db().ExecContext(ctx, `INSERT OR REPLACE INTO gogo_migrations(app, name, applied, checksum, executor_version) VALUES (?, ?, ?, ?, ?)`, migration.AppLabel, migration.Name, time.Now().UTC(), checksum, r.ExecutorVersion)
	return err
}

// RecordUnapplied deletes an applied migration row.
func (r Recorder) RecordUnapplied(ctx context.Context, dependency Dependency) error {
	_, err := r.db().ExecContext(ctx, `DELETE FROM gogo_migrations WHERE app = ? AND name = ?`, dependency.AppLabel, dependency.Name)
	return err
}

// Applied returns applied migration history.
func (r Recorder) Applied(ctx context.Context) ([]AppliedMigration, error) {
	rows, err := r.db().QueryContext(ctx, `SELECT app, name, applied, checksum, executor_version FROM gogo_migrations ORDER BY app, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var applied []AppliedMigration
	for rows.Next() {
		var item AppliedMigration
		if err := rows.Scan(&item.AppLabel, &item.Name, &item.Applied, &item.Checksum, &item.ExecutorVersion); err != nil {
			return nil, err
		}
		applied = append(applied, item)
	}
	return applied, rows.Err()
}

// CheckConsistency verifies all applied rows exist in known migrations.
func (r Recorder) CheckConsistency(ctx context.Context, known []Migration) error {
	applied, err := r.Applied(ctx)
	if err != nil {
		return err
	}
	knownSet := map[string]bool{}
	for _, migration := range known {
		knownSet[migration.Identity()] = true
	}
	for _, item := range applied {
		key := item.AppLabel + "." + item.Name
		if !knownSet[key] {
			return fmt.Errorf("%w: %s", ErrInconsistentMigrationHistory, key)
		}
	}
	return nil
}

func (r Recorder) db() *sql.DB {
	return r.Database.SQLDB()
}
