package migrations

import "testing"

func TestAutodetectorFindsModelFieldIndexConstraintChanges(t *testing.T) {
	from := NewProjectState()
	from.AddModel(ModelState{AppLabel: "blog", Name: "Post", TableName: "blog_post", Fields: []FieldState{{Name: "title", Kind: "text"}}, Indexes: []IndexState{{Name: "idx_title", Fields: []string{"title"}}}, Constraints: []ConstraintState{{Name: "uniq_title", Type: "unique", Fields: []string{"title"}}}})
	to := NewProjectState()
	to.AddModel(ModelState{AppLabel: "blog", Name: "Article", TableName: "blog_article"})
	to.AddModel(ModelState{AppLabel: "blog", Name: "Post", TableName: "blog_post", Fields: []FieldState{{Name: "headline", Kind: "text"}, {Name: "published", Kind: "boolean"}}, Indexes: []IndexState{{Name: "idx_headline", Fields: []string{"headline"}}}, Constraints: []ConstraintState{{Name: "uniq_headline", Type: "unique", Fields: []string{"headline"}}}, Options: map[string]any{"ordering": []string{"headline"}}})

	changes := NewAutodetector(from, to).Changes()
	want := map[ChangeType]bool{
		ChangeCreateModel:      true,
		ChangeRemoveField:      true,
		ChangeAddField:         true,
		ChangeAlterModel:       true,
		ChangeRemoveIndex:      true,
		ChangeAddIndex:         true,
		ChangeRemoveConstraint: true,
		ChangeAddConstraint:    true,
	}
	for _, change := range changes {
		want[change.Type] = false
	}
	for changeType, missing := range want {
		if missing {
			t.Fatalf("missing change type %s in %#v", changeType, changes)
		}
	}
}

func TestAutodetectorRenameQuestionAndMerge(t *testing.T) {
	from := NewProjectState()
	from.AddModel(ModelState{AppLabel: "blog", Name: "Post", TableName: "blog_post"})
	to := NewProjectState()
	to.AddModel(ModelState{AppLabel: "blog", Name: "Article", TableName: "blog_post"})

	detector := NewAutodetector(from, to)
	detector.Questioner = RenameQuestionerFunc(func(oldModel, newModel ModelState) bool {
		return oldModel.TableName == newModel.TableName
	})
	changes := detector.Changes()
	if len(changes) != 1 || changes[0].Type != ChangeRenameModel || changes[0].OldName != "Post" || changes[0].NewName != "Article" {
		t.Fatalf("rename changes = %#v", changes)
	}

	manual := Migration{AppLabel: "blog", Name: "0002_manual", Operations: []Operation{ManifestOperation{NameValue: "manual"}}}
	merged := MergeAutodetectedOperations(manual, []Operation{ManifestOperation{NameValue: "auto"}})
	if len(merged.Operations) != 2 || merged.Operations[0].Name() != "manual" || merged.Operations[1].Name() != "auto" {
		t.Fatalf("merged operations = %#v", merged.Operations)
	}
}
