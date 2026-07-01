package migrations

import (
	"encoding/json"
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
			{Name: "title", Column: "title", Kind: "char", ColumnTypes: map[string]string{"postgres": "varchar(200)"}, Null: true, Unique: true, DBIndex: true, DBDefault: models.DefaultValue("untitled"), DBCollation: "en_US"},
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
	title := model.Fields[1]
	if title.Kind != "char" || title.ColumnTypes["postgres"] != "varchar(200)" || !title.Null || !title.Unique || !title.DBIndex || title.DBDefault == nil || title.DBDefault.Kind != models.DefaultLiteral || title.DBDefault.Value != "untitled" || title.DBCollation != "en_US" {
		t.Fatalf("rich field state was not preserved: %#v", title)
	}
	title.ColumnTypes["postgres"] = "text"
	again := StateFromRegistry(registry).Models["blog.Post"].Fields[1]
	if again.ColumnTypes["postgres"] != "varchar(200)" {
		t.Fatalf("field ColumnTypes state was not cloned: %#v", again.ColumnTypes)
	}
}

func TestFieldStateDatabaseDefaultManifestCompatibility(t *testing.T) {
	var legacy FieldState
	if err := json.Unmarshal([]byte(`{"name":"status","db_default":"draft"}`), &legacy); err != nil {
		t.Fatalf("legacy default unmarshal error = %v", err)
	}
	if legacy.DBDefault == nil || legacy.DBDefault.Kind != models.DefaultLiteral || legacy.DBDefault.Value != "draft" {
		t.Fatalf("legacy default = %#v", legacy.DBDefault)
	}

	var expression FieldState
	if err := json.Unmarshal([]byte(`{"name":"id","db_default":{"kind":"expression","sql":"gen_random_uuid()"}}`), &expression); err != nil {
		t.Fatalf("expression default unmarshal error = %v", err)
	}
	if expression.DBDefault == nil || expression.DBDefault.Kind != models.DefaultExpression || expression.DBDefault.SQL != "gen_random_uuid()" {
		t.Fatalf("expression default = %#v", expression.DBDefault)
	}

	data, err := json.Marshal(FieldState{Name: "status", DBDefault: databaseDefaultPtr(models.DefaultValue("draft"))})
	if err != nil {
		t.Fatalf("marshal default error = %v", err)
	}
	if string(data) != `{"name":"status","db_default":{"kind":"literal","value":"draft"}}` {
		t.Fatalf("marshaled default = %s", data)
	}
}

func TestProjectStateFromRegistrySkipsUnmanagedModels(t *testing.T) {
	managed := false
	registry := models.NewRegistry()
	if err := registry.RegisterMetadata(models.Metadata{
		AppLabel:  "legacy",
		ModelName: "Order",
		TableName: "legacy_order",
		Managed:   &managed,
		Fields:    []models.FieldMeta{{Name: "id", Column: "id", PrimaryKey: true}},
	}); err != nil {
		t.Fatalf("RegisterMetadata() error = %v", err)
	}

	state := StateFromRegistry(registry)
	if _, exists := state.Models["legacy.Order"]; exists {
		t.Fatalf("unmanaged model was included in migration state: %#v", state.Models)
	}
}

func databaseDefaultPtr(defaultValue models.DatabaseDefault) *models.DatabaseDefault {
	return &defaultValue
}
