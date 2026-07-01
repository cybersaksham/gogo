package models

import (
	"errors"
	"reflect"
	"testing"
)

type registryArticle struct {
	BaseModel
}

func (registryArticle) ModelMeta() Metadata {
	return Metadata{AppLabel: "blog", ModelName: "Article"}
}

type registryComment struct {
	BaseModel
}

func (registryComment) ModelMeta() Metadata {
	return Metadata{AppLabel: "blog", ModelName: "Comment"}
}

func TestModelRegistryRegistersLooksUpAndPreservesOrder(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(registryArticle{}); err != nil {
		t.Fatalf("Register(article) error = %v", err)
	}
	if err := registry.Register(registryComment{}); err != nil {
		t.Fatalf("Register(comment) error = %v", err)
	}

	meta, ok := registry.Lookup("blog.Article")
	if !ok {
		t.Fatalf("Lookup(blog.Article) missing")
	}
	if meta.TableName != "blog_article" {
		t.Fatalf("TableName = %q, want blog_article", meta.TableName)
	}

	names := modelNames(registry.Models())
	if !reflect.DeepEqual(names, []string{"Article", "Comment"}) {
		t.Fatalf("model order = %#v, want Article, Comment", names)
	}
}

func TestModelRegistryRejectsDuplicateModelWithinApp(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(registryArticle{}); err != nil {
		t.Fatalf("Register(article) error = %v", err)
	}
	err := registry.Register(registryArticle{})
	if !errors.Is(err, ErrDuplicateModel) {
		t.Fatalf("Register(duplicate) error = %v, want ErrDuplicateModel", err)
	}
}

func TestModelRegistryReturnsImmutableMetadata(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(registryArticle{}); err != nil {
		t.Fatalf("Register(article) error = %v", err)
	}

	models := registry.Models()
	models[0].TableName = "changed"

	meta, ok := registry.Lookup("blog.Article")
	if !ok {
		t.Fatalf("Lookup(blog.Article) missing")
	}
	if meta.TableName != "blog_article" {
		t.Fatalf("stored TableName = %q, want blog_article", meta.TableName)
	}
}

func TestModelRegistryExposesMetadataForFrameworkConsumers(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(registryArticle{}); err != nil {
		t.Fatalf("Register(article) error = %v", err)
	}

	views := [][]Metadata{
		registry.AppMetadata(),
		registry.MigrationMetadata(),
		registry.ORMMetadata(),
		registry.AdminMetadata(),
		registry.SerializerMetadata(),
		registry.ContentTypeMetadata(),
	}
	for _, view := range views {
		if len(view) != 1 || view[0].ModelName != "Article" {
			t.Fatalf("metadata view = %#v, want Article", view)
		}
		view[0].ModelName = "Changed"
	}

	if got := registry.ContentTypeMetadata()[0].ModelName; got != "Article" {
		t.Fatalf("ContentTypeMetadata mutated to %q, want Article", got)
	}
}

func TestModelRegistryValidatesRelationTargets(t *testing.T) {
	registry := NewRegistry()
	if err := registry.RegisterMetadata(Metadata{
		AppLabel:  "blog",
		ModelName: "Article",
		Fields:    []FieldMeta{{Name: "id", Column: "id", PrimaryKey: true}},
	}); err != nil {
		t.Fatalf("RegisterMetadata(article) error = %v", err)
	}
	if err := registry.RegisterMetadata(Metadata{
		AppLabel:  "blog",
		ModelName: "Comment",
		Fields: []FieldMeta{
			{Name: "id", Column: "id", PrimaryKey: true},
			{Name: "article_id", Column: "article_id", RelationTarget: "blog.Article"},
			{Name: "parent_id", Column: "parent_id", RelationTarget: "self"},
		},
	}); err != nil {
		t.Fatalf("RegisterMetadata(comment) error = %v", err)
	}
	if err := registry.ValidateRelations(); err != nil {
		t.Fatalf("ValidateRelations() error = %v", err)
	}

	missing := NewRegistry()
	if err := missing.RegisterMetadata(Metadata{
		AppLabel:  "blog",
		ModelName: "Comment",
		Fields: []FieldMeta{
			{Name: "id", Column: "id", PrimaryKey: true},
			{Name: "author_id", Column: "author_id", RelationTarget: "auth.User"},
		},
	}); err != nil {
		t.Fatalf("RegisterMetadata(comment) error = %v", err)
	}
	if err := missing.ValidateRelations(); !errors.Is(err, ErrInvalidMetadata) {
		t.Fatalf("ValidateRelations() error = %v, want ErrInvalidMetadata", err)
	}
}

func modelNames(models []Metadata) []string {
	names := make([]string, len(models))
	for i, meta := range models {
		names[i] = meta.ModelName
	}
	return names
}
