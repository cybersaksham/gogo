package migrations

import (
	"context"
	"testing"
)

func TestOperationInterface(t *testing.T) {
	op := FakeOperation{NameValue: "AddField", Model: "blog.Post", Field: "title", ReversibleValue: true}
	state := NewProjectState()
	editor := &FakeSchemaEditor{}

	if op.Name() != "AddField" || op.Describe() == "" || !op.Reversible() {
		t.Fatalf("operation metadata mismatch")
	}
	if err := op.StateForwards(&state); err != nil {
		t.Fatalf("StateForwards() error = %v", err)
	}
	if err := op.DatabaseForwards(context.Background(), editor); err != nil {
		t.Fatalf("DatabaseForwards() error = %v", err)
	}
	if err := op.DatabaseBackwards(context.Background(), editor); err != nil {
		t.Fatalf("DatabaseBackwards() error = %v", err)
	}
	if !op.ReferencesModel("blog", "Post") || !op.ReferencesField("blog", "Post", "title") {
		t.Fatalf("reference checks failed")
	}
	if len(editor.SQL) != 2 {
		t.Fatalf("editor SQL calls = %#v", editor.SQL)
	}
}

type FakeOperation struct {
	NameValue       string
	Model           string
	Field           string
	ReversibleValue bool
}

func (o FakeOperation) Name() string { return o.NameValue }
func (o FakeOperation) StateForwards(*ProjectState) error {
	return nil
}
func (o FakeOperation) DatabaseForwards(ctx context.Context, editor SchemaEditor) error {
	return editor.Execute(ctx, "-- forwards")
}
func (o FakeOperation) DatabaseBackwards(ctx context.Context, editor SchemaEditor) error {
	return editor.Execute(ctx, "-- backwards")
}
func (o FakeOperation) Describe() string { return "fake operation" }
func (o FakeOperation) Reversible() bool { return o.ReversibleValue }
func (o FakeOperation) ReferencesModel(appLabel, modelName string) bool {
	return o.Model == appLabel+"."+modelName
}
func (o FakeOperation) ReferencesField(appLabel, modelName, fieldName string) bool {
	return o.ReferencesModel(appLabel, modelName) && o.Field == fieldName
}

type FakeSchemaEditor struct {
	SQL []string
}

func (e *FakeSchemaEditor) Execute(ctx context.Context, sql string, args ...any) error {
	_ = ctx
	_ = args
	e.SQL = append(e.SQL, sql)
	return nil
}
