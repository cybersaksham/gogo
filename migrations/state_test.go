package migrations

import (
	"testing"

	"github.com/cybersaksham/gogo/models"
)

func TestProjectStateCloneAndMutation(t *testing.T) {
	state := NewProjectState()
	state.AddModel(ModelState{
		AppLabel:  "blog",
		Name:      "Post",
		TableName: "blog_post",
		Fields:    []FieldState{{Name: "id", Column: "id", PrimaryKey: true}},
	})
	cloned := state.Clone()
	cloned.AddField("blog", "Post", FieldState{Name: "title", Column: "title"})
	clonedModel := cloned.Models["blog.Post"]
	clonedModel.Indexes = append(clonedModel.Indexes, IndexState{Name: "idx_title", Fields: []string{"title"}})
	cloned.Models["blog.Post"] = clonedModel

	if len(state.Models["blog.Post"].Fields) != 1 || len(state.Models["blog.Post"].Indexes) != 0 {
		t.Fatalf("original state was mutated: %#v", state.Models["blog.Post"])
	}
	if len(cloned.Models["blog.Post"].Fields) != 2 {
		t.Fatalf("cloned state missing mutation: %#v", cloned.Models["blog.Post"])
	}
}

func TestProjectStateFromRegistry(t *testing.T) {
	registry := models.NewRegistry()
	err := registry.RegisterMetadata(models.Metadata{
		AppLabel:  "blog",
		ModelName: "Post",
		TableName: "blog_post",
		Fields: []models.FieldMeta{
			{Name: "id", Column: "id", PrimaryKey: true},
			{Name: "title", Column: "title"},
		},
		Indexes:     []models.Index{{Name: "idx_title", Fields: []models.IndexField{models.Asc("title")}}},
		Constraints: []models.Constraint{{Name: "uniq_title", Type: models.ConstraintUnique, Fields: []models.IndexField{models.Asc("title")}}},
	})
	if err != nil {
		t.Fatalf("RegisterMetadata() error = %v", err)
	}

	state := StateFromRegistry(registry)
	model := state.Models["blog.Post"]
	if model.TableName != "blog_post" || len(model.Fields) != 2 || len(model.Indexes) != 1 || len(model.Constraints) != 1 {
		t.Fatalf("registry state model = %#v", model)
	}
	if model.Fields[0].Name != "id" || !model.Fields[0].PrimaryKey {
		t.Fatalf("field state = %#v", model.Fields)
	}
}
