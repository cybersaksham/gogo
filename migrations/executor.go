package migrations

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

// ExecutorOptions configure migration execution.
type ExecutorOptions struct {
	Fake        bool
	FakeInitial bool
	Plan        bool
}

// Executor applies and unapplies migrations.
type Executor struct {
	Recorder Recorder
	Editor   SchemaEditor
}

// NewExecutor creates a migration executor.
func NewExecutor(recorder Recorder, editor SchemaEditor) Executor {
	return Executor{Recorder: recorder, Editor: editor}
}

// Apply applies migrations in order.
func (e Executor) Apply(ctx context.Context, migrations []Migration, options ExecutorOptions) (err error) {
	release, err := e.lock(ctx, options)
	if err != nil {
		return err
	}
	defer func() {
		if release != nil {
			if releaseErr := release(context.Background()); err == nil {
				err = releaseErr
			}
		}
	}()
	if err := e.Recorder.EnsureSchema(ctx); err != nil {
		return err
	}
	for _, migration := range migrations {
		applied, err := e.Recorder.IsApplied(ctx, migration.Dependency())
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		fakeInitial, err := e.shouldFakeInitial(ctx, migration, options)
		if err != nil {
			return err
		}
		if !options.Fake && !fakeInitial && !options.Plan {
			for _, operation := range migration.Operations {
				if err := operation.DatabaseForwards(ctx, e.Editor); err != nil {
					return err
				}
			}
		}
		if !options.Plan {
			if err := e.Recorder.RecordApplied(ctx, migration, migrationChecksum(migration)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Rollback unapplies one migration.
func (e Executor) Rollback(ctx context.Context, migration Migration, options ExecutorOptions) (err error) {
	release, err := e.lock(ctx, options)
	if err != nil {
		return err
	}
	defer func() {
		if release != nil {
			if releaseErr := release(context.Background()); err == nil {
				err = releaseErr
			}
		}
	}()
	for i := len(migration.Operations) - 1; i >= 0; i-- {
		operation := migration.Operations[i]
		if !operation.Reversible() {
			return ErrIrreversibleOperation
		}
		if !options.Fake && !options.Plan {
			if err := operation.DatabaseBackwards(ctx, e.Editor); err != nil {
				return err
			}
		}
	}
	if !options.Plan {
		return e.Recorder.RecordUnapplied(ctx, migration.Dependency())
	}
	return nil
}

func (e Executor) lock(ctx context.Context, options ExecutorOptions) (func(context.Context) error, error) {
	if options.Plan {
		return nil, nil
	}
	return e.Recorder.AcquireLock(ctx)
}

func (e Executor) shouldFakeInitial(ctx context.Context, migration Migration, options ExecutorOptions) (bool, error) {
	if !options.FakeInitial || !isInitialMigration(migration) {
		return false, nil
	}
	checker, ok := e.Editor.(TableExistenceChecker)
	if !ok {
		return false, nil
	}
	tables := initialMigrationTables(migration)
	if len(tables) == 0 {
		return false, nil
	}
	for _, table := range tables {
		exists, err := checker.TableExists(ctx, table)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return true, nil
}

func isInitialMigration(migration Migration) bool {
	return migration.Name == InitialMigrationName() || strings.HasPrefix(migration.Name, "0001_")
}

func initialMigrationTables(migration Migration) []string {
	seen := map[string]struct{}{}
	var tables []string
	for _, operation := range migration.Operations {
		provider, ok := operation.(InitialTableProvider)
		if !ok {
			continue
		}
		for _, table := range provider.InitialTables() {
			if table == "" {
				continue
			}
			if _, exists := seen[table]; exists {
				continue
			}
			seen[table] = struct{}{}
			tables = append(tables, table)
		}
	}
	return tables
}

func migrationChecksum(migration Migration) string {
	hash := sha1.New()
	hash.Write([]byte(migration.Identity()))
	for _, operation := range migration.Operations {
		hash.Write([]byte("|" + operation.Name()))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
