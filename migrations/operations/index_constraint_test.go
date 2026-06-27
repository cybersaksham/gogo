package operations

import (
	"context"
	"testing"

	"github.com/cybersaksham/gogo/migrations"
)

func TestIndexAndConstraintOperations(t *testing.T) {
	state := migrations.NewProjectState()
	state.AddModel(migrations.ModelState{AppLabel: "blog", Name: "Post", TableName: "blog_post"})
	editor := &fakeEditor{}

	addIndex := AddIndex{AppLabel: "blog", ModelName: "Post", Index: migrations.IndexState{Name: "idx_title", Fields: []string{"title"}}}
	if err := addIndex.StateForwards(&state); err != nil {
		t.Fatalf("AddIndex StateForwards() error = %v", err)
	}
	if err := addIndex.DatabaseForwards(context.Background(), editor); err != nil {
		t.Fatalf("AddIndex DatabaseForwards() error = %v", err)
	}
	if len(state.Models["blog.Post"].Indexes) != 1 || editor.SQL[0] != `CREATE INDEX idx_title ON blog_post (title)` {
		t.Fatalf("index state/sql = %#v / %#v", state.Models["blog.Post"].Indexes, editor.SQL)
	}

	renameIndex := RenameIndex{AppLabel: "blog", ModelName: "Post", OldName: "idx_title", NewName: "idx_headline"}
	if err := renameIndex.StateForwards(&state); err != nil {
		t.Fatalf("RenameIndex StateForwards() error = %v", err)
	}
	if state.Models["blog.Post"].Indexes[0].Name != "idx_headline" {
		t.Fatalf("renamed index = %#v", state.Models["blog.Post"].Indexes[0])
	}

	addConstraint := AddConstraint{AppLabel: "blog", ModelName: "Post", Constraint: migrations.ConstraintState{Name: "uniq_title", Type: "unique", Fields: []string{"title"}}}
	if err := addConstraint.StateForwards(&state); err != nil {
		t.Fatalf("AddConstraint StateForwards() error = %v", err)
	}
	if err := addConstraint.DatabaseForwards(context.Background(), editor); err != nil {
		t.Fatalf("AddConstraint DatabaseForwards() error = %v", err)
	}
	if len(state.Models["blog.Post"].Constraints) != 1 || editor.SQL[len(editor.SQL)-1] != `ALTER TABLE blog_post ADD CONSTRAINT uniq_title UNIQUE (title)` {
		t.Fatalf("constraint state/sql = %#v / %#v", state.Models["blog.Post"].Constraints, editor.SQL)
	}

	if err := (RemoveIndex{AppLabel: "blog", ModelName: "Post", IndexName: "idx_headline"}).StateForwards(&state); err != nil {
		t.Fatalf("RemoveIndex StateForwards() error = %v", err)
	}
	if err := (RemoveConstraint{AppLabel: "blog", ModelName: "Post", ConstraintName: "uniq_title"}).StateForwards(&state); err != nil {
		t.Fatalf("RemoveConstraint StateForwards() error = %v", err)
	}
	if len(state.Models["blog.Post"].Indexes) != 0 || len(state.Models["blog.Post"].Constraints) != 0 {
		t.Fatalf("remove state = %#v", state.Models["blog.Post"])
	}
}
