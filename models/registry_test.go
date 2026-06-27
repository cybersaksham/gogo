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

func modelNames(models []Metadata) []string {
	names := make([]string, len(models))
	for i, meta := range models {
		names[i] = meta.ModelName
	}
	return names
}
