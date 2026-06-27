package migrations

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
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
func (e Executor) Apply(ctx context.Context, migrations []Migration, options ExecutorOptions) error {
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
		if !options.Fake && !options.Plan {
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
func (e Executor) Rollback(ctx context.Context, migration Migration, options ExecutorOptions) error {
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

func migrationChecksum(migration Migration) string {
	hash := sha1.New()
	hash.Write([]byte(migration.Identity()))
	for _, operation := range migration.Operations {
		hash.Write([]byte("|" + operation.Name()))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
