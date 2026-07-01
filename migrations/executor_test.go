package migrations

import (
	"context"
	"errors"
	"testing"
)

func TestExecutorApplyRollbackAndFake(t *testing.T) {
	ctx := context.Background()
	db := openRecorderDB(t)
	defer db.Close()
	recorder := NewRecorder(db, "executor")
	editor := &FakeSchemaEditor{}
	executor := NewExecutor(recorder, editor)
	migration := testMigration("blog", "0001_initial")
	migration.Operations = []Operation{FakeOperation{NameValue: "op", ReversibleValue: true}}

	if err := executor.Apply(ctx, []Migration{migration}, ExecutorOptions{}); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if ok, _ := recorder.IsApplied(ctx, migration.Dependency()); !ok {
		t.Fatalf("migration not recorded applied")
	}
	if err := executor.Rollback(ctx, migration, ExecutorOptions{}); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}
	if ok, _ := recorder.IsApplied(ctx, migration.Dependency()); ok {
		t.Fatalf("migration not recorded unapplied")
	}

	if err := executor.Apply(ctx, []Migration{migration}, ExecutorOptions{Fake: true}); err != nil {
		t.Fatalf("fake Apply() error = %v", err)
	}
	if len(editor.SQL) != 2 {
		t.Fatalf("fake apply executed SQL: %#v", editor.SQL)
	}
	if err := executor.Rollback(ctx, migration, ExecutorOptions{Fake: true}); err != nil {
		t.Fatalf("fake Rollback() error = %v", err)
	}
}

func TestExecutorFailureAndIrreversibleReverse(t *testing.T) {
	ctx := context.Background()
	db := openRecorderDB(t)
	defer db.Close()
	recorder := NewRecorder(db, "executor")
	executor := NewExecutor(recorder, &FakeSchemaEditor{})
	failing := testMigration("blog", "0002_fail")
	failing.Atomic = true
	failing.Operations = []Operation{FailingOperation{}}
	if err := executor.Apply(ctx, []Migration{failing}, ExecutorOptions{}); err == nil {
		t.Fatalf("failing migration returned nil")
	}
	if ok, _ := recorder.IsApplied(ctx, failing.Dependency()); ok {
		t.Fatalf("failing migration was recorded")
	}

	irreversible := testMigration("blog", "0003_irreversible")
	irreversible.Operations = []Operation{FakeOperation{NameValue: "irreversible", ReversibleValue: false}}
	if err := executor.Rollback(ctx, irreversible, ExecutorOptions{}); !errors.Is(err, ErrIrreversibleOperation) {
		t.Fatalf("Rollback irreversible error = %v", err)
	}
}

func TestExecutorApplyRequiresMigrationLock(t *testing.T) {
	ctx := context.Background()
	db := openRecorderDB(t)
	defer db.Close()
	recorder := NewRecorder(db, "executor")
	release, err := recorder.AcquireLock(ctx)
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}
	defer func() {
		if err := release(ctx); err != nil {
			t.Fatalf("release() error = %v", err)
		}
	}()

	editor := &FakeSchemaEditor{}
	executor := NewExecutor(recorder, editor)
	migration := testMigration("blog", "0001_initial")
	migration.Operations = []Operation{FakeOperation{NameValue: "op", ReversibleValue: true}}
	if err := executor.Apply(ctx, []Migration{migration}, ExecutorOptions{}); !errors.Is(err, ErrMigrationLocked) {
		t.Fatalf("Apply() error = %v, want ErrMigrationLocked", err)
	}
	if len(editor.SQL) != 0 {
		t.Fatalf("locked Apply() executed SQL: %#v", editor.SQL)
	}
}

type FailingOperation struct{}

func (FailingOperation) Name() string                      { return "FailingOperation" }
func (FailingOperation) StateForwards(*ProjectState) error { return nil }
func (FailingOperation) DatabaseForwards(context.Context, SchemaEditor) error {
	return errors.New("failed")
}
func (FailingOperation) DatabaseBackwards(context.Context, SchemaEditor) error { return nil }
func (FailingOperation) Describe() string                                      { return "failing" }
func (FailingOperation) Reversible() bool                                      { return true }
func (FailingOperation) ReferencesModel(string, string) bool                   { return false }
func (FailingOperation) ReferencesField(string, string, string) bool           { return false }
