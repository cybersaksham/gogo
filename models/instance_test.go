package models

import (
	"context"
	"reflect"
	"testing"
)

type instanceModel struct {
	BaseModel
	calls []string
}

func (m *instanceModel) ModelMeta() Metadata {
	return Metadata{AppLabel: "tests", ModelName: "Instance"}
}

func (m *instanceModel) CleanFields(context.Context) error {
	m.calls = append(m.calls, "clean-fields")
	return nil
}

func (m *instanceModel) Clean(context.Context) error {
	m.calls = append(m.calls, "clean")
	return nil
}

func (m *instanceModel) ValidateUnique(context.Context) error {
	m.calls = append(m.calls, "validate-unique")
	return nil
}

func (m *instanceModel) ValidateConstraints(context.Context) error {
	m.calls = append(m.calls, "validate-constraints")
	return nil
}

func (m *instanceModel) AbsoluteURL() string {
	return "/instances/1/"
}

func (m *instanceModel) FieldDisplay(field string) (string, bool) {
	if field == "status" {
		return "Published", true
	}
	return "", false
}

func (m *instanceModel) SerializableValue(field string) (any, bool) {
	if field == "id" {
		return int64(1), true
	}
	return nil, false
}

func (m *instanceModel) NaturalKey() []any {
	return []any{"instance", int64(1)}
}

type recordingStore struct {
	saveModel     Model
	saveOptions   SaveOptions
	deleteModel   Model
	deleteOptions DeleteOptions
	refreshed     bool
	fromDB        map[string]any
}

func (s *recordingStore) Save(_ context.Context, model Model, options SaveOptions) error {
	s.saveModel = model
	s.saveOptions = options
	return nil
}

func (s *recordingStore) Delete(_ context.Context, model Model, options DeleteOptions) error {
	s.deleteModel = model
	s.deleteOptions = options
	return nil
}

func (s *recordingStore) RefreshFromDB(context.Context, Model, RefreshOptions) error {
	s.refreshed = true
	return nil
}

func (s *recordingStore) FromDB(_ context.Context, _ Model, values map[string]any) error {
	s.fromDB = values
	return nil
}

func TestSavePassesOptionsAndMarksLoaded(t *testing.T) {
	model := &instanceModel{}
	store := &recordingStore{}

	err := Save(context.Background(), model, store,
		ForceInsert(),
		ForceUpdate(),
		UpdateFields("name", "status"),
		UsingDatabase("replica"),
		RawSave(),
	)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if store.saveModel != model {
		t.Fatalf("save model = %#v, want model", store.saveModel)
	}
	if !store.saveOptions.ForceInsert || !store.saveOptions.ForceUpdate || !store.saveOptions.Raw {
		t.Fatalf("save options = %#v, want force insert/update and raw", store.saveOptions)
	}
	if !reflect.DeepEqual(store.saveOptions.UpdateFields, []string{"name", "status"}) {
		t.Fatalf("UpdateFields = %#v, want name/status", store.saveOptions.UpdateFields)
	}
	if store.saveOptions.Using != "replica" {
		t.Fatalf("Using = %q, want replica", store.saveOptions.Using)
	}
	if model.ModelState() != StateLoaded {
		t.Fatalf("ModelState() = %s, want loaded", model.ModelState())
	}
}

func TestDeletePassesOptionsAndMarksDeleted(t *testing.T) {
	model := &instanceModel{}
	store := &recordingStore{}

	err := Delete(context.Background(), model, store, DeleteUsingDatabase("archive"), KeepParentRows())
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.deleteModel != model {
		t.Fatalf("delete model = %#v, want model", store.deleteModel)
	}
	if store.deleteOptions.Using != "archive" || !store.deleteOptions.KeepParents {
		t.Fatalf("delete options = %#v, want archive and keep parents", store.deleteOptions)
	}
	if model.ModelState() != StateDeleted {
		t.Fatalf("ModelState() = %s, want deleted", model.ModelState())
	}
}

func TestFullCleanRunsAllValidationSteps(t *testing.T) {
	model := &instanceModel{}

	if err := FullClean(context.Background(), model); err != nil {
		t.Fatalf("FullClean() error = %v", err)
	}

	want := []string{"clean-fields", "clean", "validate-unique", "validate-constraints"}
	if !reflect.DeepEqual(model.calls, want) {
		t.Fatalf("calls = %#v, want %#v", model.calls, want)
	}
}

func TestRefreshAndFromDBMarkLoaded(t *testing.T) {
	model := &instanceModel{}
	store := &recordingStore{}

	if err := RefreshFromDB(context.Background(), model, store, RefreshOptions{Using: "replica", Fields: []string{"name"}}); err != nil {
		t.Fatalf("RefreshFromDB() error = %v", err)
	}
	if !store.refreshed || model.ModelState() != StateLoaded {
		t.Fatalf("refresh = %v state = %s, want refreshed loaded", store.refreshed, model.ModelState())
	}

	if err := FromDB(context.Background(), model, store, map[string]any{"id": int64(1)}); err != nil {
		t.Fatalf("FromDB() error = %v", err)
	}
	if store.fromDB["id"] != int64(1) || model.ModelState() != StateLoaded {
		t.Fatalf("fromDB = %#v state = %s, want loaded id", store.fromDB, model.ModelState())
	}
}

func TestInstanceDisplayAndIdentityHelpers(t *testing.T) {
	model := &instanceModel{}

	if got := GetAbsoluteURL(model); got != "/instances/1/" {
		t.Fatalf("GetAbsoluteURL() = %q, want /instances/1/", got)
	}
	if got, ok := GetFieldDisplay(model, "status"); !ok || got != "Published" {
		t.Fatalf("GetFieldDisplay() = (%q, %v), want Published true", got, ok)
	}
	if got, ok := SerializableValue(model, "id"); !ok || got != int64(1) {
		t.Fatalf("SerializableValue() = (%#v, %v), want 1 true", got, ok)
	}
	if got := NaturalKey(model); !reflect.DeepEqual(got, []any{"instance", int64(1)}) {
		t.Fatalf("NaturalKey() = %#v, want instance key", got)
	}
}
