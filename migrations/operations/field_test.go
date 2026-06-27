package operations

import (
	"context"
	"errors"
	"testing"

	"github.com/cybersaksham/gogo/migrations"
)

func TestFieldOperationsMutateStateAndRenderSQL(t *testing.T) {
	state := migrations.NewProjectState()
	state.AddModel(migrations.ModelState{AppLabel: "blog", Name: "Post", TableName: "blog_post"})
	editor := &fakeEditor{}

	add := AddField{AppLabel: "blog", ModelName: "Post", Field: migrations.FieldState{Name: "title", Column: "title", Kind: "text", Null: true}}
	if err := add.StateForwards(&state); err != nil {
		t.Fatalf("AddField StateForwards() error = %v", err)
	}
	if err := add.DatabaseForwards(context.Background(), editor); err != nil {
		t.Fatalf("AddField DatabaseForwards() error = %v", err)
	}
	if state.Models["blog.Post"].Fields[0].Name != "title" || editor.SQL[0] != `ALTER TABLE blog_post ADD COLUMN title text` {
		t.Fatalf("add field state/sql = %#v / %#v", state.Models["blog.Post"].Fields, editor.SQL)
	}

	alter := AlterField{AppLabel: "blog", ModelName: "Post", OldField: migrations.FieldState{Name: "title", Column: "title", Kind: "text", Null: true}, NewField: migrations.FieldState{Name: "title", Column: "title", Kind: "varchar(255)", Null: false}}
	if err := alter.StateForwards(&state); err != nil {
		t.Fatalf("AlterField StateForwards() error = %v", err)
	}
	if state.Models["blog.Post"].Fields[0].Kind != "varchar(255)" || state.Models["blog.Post"].Fields[0].Null {
		t.Fatalf("altered field = %#v", state.Models["blog.Post"].Fields[0])
	}

	rename := RenameField{AppLabel: "blog", ModelName: "Post", OldName: "title", NewName: "headline"}
	if err := rename.StateForwards(&state); err != nil {
		t.Fatalf("RenameField StateForwards() error = %v", err)
	}
	if state.Models["blog.Post"].Fields[0].Name != "headline" {
		t.Fatalf("renamed field = %#v", state.Models["blog.Post"].Fields[0])
	}

	remove := RemoveField{AppLabel: "blog", ModelName: "Post", Field: state.Models["blog.Post"].Fields[0]}
	if err := remove.StateForwards(&state); err != nil {
		t.Fatalf("RemoveField StateForwards() error = %v", err)
	}
	if len(state.Models["blog.Post"].Fields) != 0 {
		t.Fatalf("field still exists: %#v", state.Models["blog.Post"].Fields)
	}
}

func TestAddFieldRejectsUnsafeNonNullWithoutDefault(t *testing.T) {
	op := AddField{AppLabel: "blog", ModelName: "Post", Field: migrations.FieldState{Name: "slug", Kind: "text", Null: false}}
	if err := op.ValidateSafety(); !errors.Is(err, migrations.ErrUnsafeMigration) {
		t.Fatalf("ValidateSafety() error = %v, want ErrUnsafeMigration", err)
	}
	op.HasDefault = true
	if err := op.ValidateSafety(); err != nil {
		t.Fatalf("ValidateSafety() with default error = %v", err)
	}
}
